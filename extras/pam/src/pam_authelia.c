/*
 * pam_authelia.c - PAM module for Authelia SSH authentication.
 *
 * This is a thin shim that handles PAM conversation (prompting) and delegates
 * all authentication logic to the authelia-pam Go binary via a stdin/stdout
 * pipe protocol.
 *
 * Copyright 2024 Authelia Contributors
 * SPDX-License-Identifier: Apache-2.0
 */

#define __STDC_WANT_LIB_EXT1__ 1
#define _DEFAULT_SOURCE

#include <errno.h>
#include <poll.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>
#include <time.h>
#include <unistd.h>

#ifdef __linux__
#include <sys/prctl.h>
#endif

#include <security/pam_appl.h>
#include <security/pam_modules.h>

#include "pam_authelia.h"

/* -------------------------------------------------------------------------- */
/* Helpers                                                                    */
/* -------------------------------------------------------------------------- */

/*
 * Securely clear memory. Uses explicit_bzero where available to prevent
 * compiler optimization from eliding the clear.
 */
static void
secure_clear(void *ptr, size_t len)
{
#if defined(__APPLE__)
	memset_s(ptr, len, 0, len);
#elif defined(__linux__) || defined(__FreeBSD__)
	explicit_bzero(ptr, len);
#else
	volatile unsigned char *p = (volatile unsigned char *)ptr;
	while (len--) {
		*p++ = 0;
	}
#endif
}

/*
 * Prompt the user via the PAM conversation function.
 * msg_style: PAM_PROMPT_ECHO_OFF, PAM_PROMPT_ECHO_ON, or PAM_TEXT_INFO.
 * If response is non-NULL, the caller must free the returned string.
 */
static int
authelia_pam_prompt(pam_handle_t *pamh, int msg_style, const char *prompt_text, char **response)
{
	const struct pam_conv *conv = NULL;
	struct pam_message msg;
	const struct pam_message *msgp = &msg;
	struct pam_response *resp = NULL;
	int ret;

	ret = pam_get_item(pamh, PAM_CONV, (const void **)&conv);
	if (ret != PAM_SUCCESS || conv == NULL || conv->conv == NULL) {
		return PAM_CONV_ERR;
	}

	memset(&msg, 0, sizeof(msg));
	msg.msg_style = msg_style;
	msg.msg = (char *)prompt_text;

	ret = conv->conv(1, &msgp, &resp, conv->appdata_ptr);
	if (ret != PAM_SUCCESS) {
		if (resp != NULL) {
			if (resp->resp != NULL) {
				secure_clear(resp->resp, strlen(resp->resp));
				free(resp->resp);
			}
			free(resp);
		}
		return ret;
	}

	if (response != NULL) {
		if (resp != NULL && resp->resp != NULL) {
			*response = resp->resp;
			resp->resp = NULL; /* Transfer ownership. */
		} else {
			*response = NULL;
		}
	}

	if (resp != NULL) {
		free(resp);
	}

	return PAM_SUCCESS;
}

/* -------------------------------------------------------------------------- */
/* Configuration                                                              */
/* -------------------------------------------------------------------------- */

static void
config_init(struct pam_authelia_config *cfg)
{
	cfg->binary              = PAM_AUTHELIA_DEFAULT_BINARY;
	cfg->url                 = NULL;
	cfg->auth_level          = "1FA+2FA";
	cfg->cookie_name         = "authelia_session";
	cfg->ca_cert             = NULL;
	cfg->method_priority     = NULL;
	cfg->oauth2_client_id    = NULL;
	cfg->oauth2_client_secret = NULL;
	cfg->oauth2_scope        = NULL;
	cfg->timeout             = PAM_AUTHELIA_DEFAULT_TIMEOUT;
	cfg->debug               = 0;
}

static int
config_parse(struct pam_authelia_config *cfg, int argc, const char **argv)
{
	int i;

	for (i = 0; i < argc; i++) {
		if (strncmp(argv[i], "url=", 4) == 0) {
			cfg->url = argv[i] + 4;
		} else if (strncmp(argv[i], "auth-level=", 11) == 0) {
			cfg->auth_level = argv[i] + 11;
		} else if (strncmp(argv[i], "cookie-name=", 12) == 0) {
			cfg->cookie_name = argv[i] + 12;
		} else if (strncmp(argv[i], "ca-cert=", 8) == 0) {
			cfg->ca_cert = argv[i] + 8;
		} else if (strncmp(argv[i], "timeout=", 8) == 0) {
			cfg->timeout = atoi(argv[i] + 8);
			if (cfg->timeout <= 0) {
				cfg->timeout = PAM_AUTHELIA_DEFAULT_TIMEOUT;
			}
		} else if (strncmp(argv[i], "binary=", 7) == 0) {
			cfg->binary = argv[i] + 7;
		} else if (strncmp(argv[i], "method-priority=", 16) == 0) {
			cfg->method_priority = argv[i] + 16;
		} else if (strncmp(argv[i], "oauth2-client-id=", 17) == 0) {
			cfg->oauth2_client_id = argv[i] + 17;
		} else if (strncmp(argv[i], "oauth2-client-secret=", 21) == 0) {
			cfg->oauth2_client_secret = argv[i] + 21;
		} else if (strncmp(argv[i], "oauth2-scope=", 13) == 0) {
			cfg->oauth2_scope = argv[i] + 13;
		} else if (strcmp(argv[i], "debug") == 0) {
			cfg->debug = 1;
		}
	}

	if (cfg->url == NULL) {
		return -1;
	}

	return 0;
}

/* -------------------------------------------------------------------------- */
/* Write a line to file descriptor with error checking.                       */
/* -------------------------------------------------------------------------- */

static int
write_line(int fd, const char *line)
{
	size_t len = strlen(line);
	ssize_t n;

	while (len > 0) {
		n = write(fd, line, len);
		if (n < 0) {
			if (errno == EINTR) continue;
			return -1;
		}
		line += n;
		len -= (size_t)n;
	}

	/* Write trailing newline. */
	while (1) {
		n = write(fd, "\n", 1);
		if (n < 0) {
			if (errno == EINTR) continue;
			return -1;
		}
		break;
	}

	return 0;
}

/* -------------------------------------------------------------------------- */
/* Read a line from a file descriptor, aborting if deadline (monotonic-time    */
/* seconds) is reached. Returns 0 on success, -1 on EOF/error, 1 on timeout.   */
/* -------------------------------------------------------------------------- */

static time_t
monotonic_seconds(void)
{
	struct timespec ts;
	if (clock_gettime(CLOCK_MONOTONIC, &ts) != 0) {
		return time(NULL);
	}
	return ts.tv_sec;
}

static int
read_line(int fd, char *buf, size_t bufsz, time_t deadline)
{
	size_t pos = 0;
	ssize_t n;
	char c;

	while (pos < bufsz - 1) {
		time_t now = monotonic_seconds();
		if (now >= deadline) {
			return 1;
		}

		struct pollfd pfd;
		pfd.fd = fd;
		pfd.events = POLLIN;

		int remaining_ms = (int)((deadline - now) * 1000);
		int pr = poll(&pfd, 1, remaining_ms);
		if (pr < 0) {
			if (errno == EINTR) continue;
			return -1;
		}
		if (pr == 0) {
			return 1; /* Timeout. */
		}

		n = read(fd, &c, 1);
		if (n < 0) {
			if (errno == EINTR) continue;
			return -1;
		}
		if (n == 0) {
			if (pos == 0) return -1;
			break;
		}
		if (c == '\n') {
			break;
		}
		buf[pos++] = c;
	}

	buf[pos] = '\0';
	return 0;
}

/* -------------------------------------------------------------------------- */
/* PAM module entry point.                                                    */
/* -------------------------------------------------------------------------- */

PAM_EXTERN int
pam_sm_authenticate(pam_handle_t *pamh, int flags, int argc, const char **argv)
{
	struct pam_authelia_config cfg;
	const char *username = NULL;
	const char *authtok = NULL;
	pid_t child;
	int pipe_to_child[2];   /* Parent writes, child reads (child's stdin).  */
	int pipe_from_child[2]; /* Child writes, parent reads (child's stdout). */
	int status;
	int ret = PAM_AUTH_ERR;
	char line[MAX_LINE];

	(void)flags;

	/* Parse configuration. */
	config_init(&cfg);
	if (config_parse(&cfg, argc, argv) != 0) {
		return PAM_AUTH_ERR;
	}

	/* Get username. */
	if (pam_get_user(pamh, &username, NULL) != PAM_SUCCESS || username == NULL) {
		return PAM_AUTH_ERR;
	}

	/* Create pipes. */
	if (pipe(pipe_to_child) != 0) {
		return PAM_AUTH_ERR;
	}
	if (pipe(pipe_from_child) != 0) {
		close(pipe_to_child[0]);
		close(pipe_to_child[1]);
		return PAM_AUTH_ERR;
	}

	child = fork();
	if (child < 0) {
		close(pipe_to_child[0]);
		close(pipe_to_child[1]);
		close(pipe_from_child[0]);
		close(pipe_from_child[1]);
		return PAM_AUTH_ERR;
	}

	if (child == 0) {
		/* ---- Child process ---- */
		close(pipe_to_child[1]);   /* Close write end of stdin pipe.  */
		close(pipe_from_child[0]); /* Close read end of stdout pipe.  */

		/* Redirect stdin/stdout. */
		if (dup2(pipe_to_child[0], STDIN_FILENO) < 0) {
			_exit(1);
		}
		if (dup2(pipe_from_child[1], STDOUT_FILENO) < 0) {
			_exit(1);
		}

		close(pipe_to_child[0]);
		close(pipe_from_child[1]);

#ifdef __linux__
		/*
		 * Ask the kernel to send SIGTERM to this child if the parent process ever
		 * dies (e.g. sshd drops the session mid-authentication). Without this, a
		 * long-running operation like OAuth2 device-authorization polling would
		 * survive past the end of the SSH session and keep hammering the API.
		 */
		prctl(PR_SET_PDEATHSIG, SIGTERM);

		/*
		 * PR_SET_PDEATHSIG is only effective while the calling task's parent has
		 * not already exited. If the parent died between fork() and now we would
		 * never be signaled, so verify the parent is still alive and exit if not.
		 */
		if (getppid() == 1) {
			_exit(1);
		}
#endif

		/* Build argument list for the Go binary. */
		char timeout_str[16];
		snprintf(timeout_str, sizeof(timeout_str), "%d", cfg.timeout);

		char *args[MAX_ARGS];
		int ai = 0;

		args[ai++] = (char *)cfg.binary;
		args[ai++] = "--url";
		args[ai++] = (char *)cfg.url;
		args[ai++] = "--auth-level";
		args[ai++] = (char *)cfg.auth_level;
		args[ai++] = "--cookie-name";
		args[ai++] = (char *)cfg.cookie_name;
		args[ai++] = "--timeout";
		args[ai++] = timeout_str;

		if (cfg.ca_cert != NULL) {
			args[ai++] = "--ca-cert";
			args[ai++] = (char *)cfg.ca_cert;
		}

		if (cfg.method_priority != NULL) {
			args[ai++] = "--method-priority";
			args[ai++] = (char *)cfg.method_priority;
		}

		if (cfg.oauth2_client_id != NULL) {
			args[ai++] = "--oauth2-client-id";
			args[ai++] = (char *)cfg.oauth2_client_id;
		}

		if (cfg.oauth2_client_secret != NULL) {
			args[ai++] = "--oauth2-client-secret";
			args[ai++] = (char *)cfg.oauth2_client_secret;
		}

		if (cfg.oauth2_scope != NULL) {
			args[ai++] = "--oauth2-scope";
			args[ai++] = (char *)cfg.oauth2_scope;
		}

		if (cfg.debug) {
			args[ai++] = "--debug";
		}

		args[ai] = NULL;

		execv(cfg.binary, args);

		/* execv only returns on error. */
		_exit(127);
	}

	/* ---- Parent process ---- */
	close(pipe_to_child[0]);   /* Close read end of stdin pipe.   */
	close(pipe_from_child[1]); /* Close write end of stdout pipe. */

	/* Send username to the Go binary. */
	if (write_line(pipe_to_child[1], username) != 0) {
		goto cleanup;
	}

	/*
	 * Send the password to the Go binary. We always try PAM_AUTHTOK first so that
	 * a preceding module (e.g. pam_unix) that already prompted for the password
	 * doesn't cause a second prompt here. If nothing in the stack has set the token,
	 * we prompt via pam_conv ourselves.
	 *
	 * When method-priority starts with "device_authorization", no password is needed
	 * (the device flow handles authentication end-to-end). We send an empty placeholder
	 * to keep the protocol in sync. The Go binary decides whether to actually skip 1FA
	 * based on the full priority list and whether oauth2-client-id is configured.
	 */
	int device_first = 0;
	if (cfg.method_priority != NULL && strncmp(cfg.method_priority, "device_authorization", 20) == 0) {
		char next = cfg.method_priority[20];
		device_first = (next == '\0' || next == ',');
	}

	if (device_first) {
		write_line(pipe_to_child[1], "");
	} else if (pam_get_item(pamh, PAM_AUTHTOK, (const void **)&authtok) == PAM_SUCCESS && authtok != NULL) {
		write_line(pipe_to_child[1], authtok);
	} else {
		char *pw = NULL;
		if (authelia_pam_prompt(pamh, PAM_PROMPT_ECHO_OFF, "Password: ", &pw) != PAM_SUCCESS || pw == NULL) {
			goto cleanup;
		}
		write_line(pipe_to_child[1], pw);
		secure_clear(pw, strlen(pw));
		free(pw);
	}

	/*
	 * Protocol loop: read commands from the Go binary until it reports SUCCESS/FAILURE
	 * or the overall timeout (cfg.timeout seconds, default 60) is reached. The deadline
	 * guards against hung children — e.g. OAuth2 device-authorization polling that
	 * outlives the SSH session, which can happen when sshd doesn't promptly signal us
	 * after a client disconnect.
	 */
	time_t deadline = monotonic_seconds() + cfg.timeout;

	while (1) {
		int rc = read_line(pipe_from_child[0], line, sizeof(line), deadline);
		if (rc != 0) {
			break;
		}

		if (strncmp(line, CMD_PROMPT_HIDDEN, strlen(CMD_PROMPT_HIDDEN)) == 0) {
			char *response = NULL;
			const char *pt = line + strlen(CMD_PROMPT_HIDDEN);

			if (authelia_pam_prompt(pamh, PAM_PROMPT_ECHO_OFF, pt, &response) != PAM_SUCCESS) {
				goto cleanup;
			}
			if (response != NULL) {
				write_line(pipe_to_child[1], response);
				secure_clear(response, strlen(response));
				free(response);
			} else {
				write_line(pipe_to_child[1], "");
			}
		} else if (strncmp(line, CMD_PROMPT_VISIBLE, strlen(CMD_PROMPT_VISIBLE)) == 0) {
			char *response = NULL;
			const char *pt = line + strlen(CMD_PROMPT_VISIBLE);

			if (authelia_pam_prompt(pamh, PAM_PROMPT_ECHO_ON, pt, &response) != PAM_SUCCESS) {
				goto cleanup;
			}
			if (response != NULL) {
				write_line(pipe_to_child[1], response);
				secure_clear(response, strlen(response));
				free(response);
			} else {
				write_line(pipe_to_child[1], "");
			}
		} else if (strncmp(line, CMD_INFO, strlen(CMD_INFO)) == 0) {
			const char *info_text = line + strlen(CMD_INFO);

			authelia_pam_prompt(pamh, PAM_TEXT_INFO, info_text, NULL);
		} else if (strcmp(line, CMD_SUCCESS) == 0) {
			ret = PAM_SUCCESS;
			break;
		} else if (strncmp(line, CMD_FAILURE, strlen(CMD_FAILURE)) == 0) {
			ret = PAM_AUTH_ERR;
			break;
		} else {
			/* Unknown command; treat as failure. */
			break;
		}
	}

cleanup:
	close(pipe_to_child[1]);
	close(pipe_from_child[0]);

	/* Wait for child or kill it on timeout. */
	if (waitpid(child, &status, WNOHANG) == 0) {
		kill(child, SIGTERM);
		waitpid(child, &status, 0);
	}

	secure_clear(line, sizeof(line));

	return ret;
}

PAM_EXTERN int
pam_sm_setcred(pam_handle_t *pamh, int flags, int argc, const char **argv)
{
	(void)pamh;
	(void)flags;
	(void)argc;
	(void)argv;

	return PAM_SUCCESS;
}
