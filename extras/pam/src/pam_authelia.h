#ifndef PAM_AUTHELIA_H
#define PAM_AUTHELIA_H

#define PAM_AUTHELIA_DEFAULT_BINARY  "/usr/local/bin/authelia-pam"
#define PAM_AUTHELIA_DEFAULT_TIMEOUT 60

/* Protocol commands from the Go binary. */
#define CMD_PROMPT_HIDDEN  "PROMPT_HIDDEN:"
#define CMD_PROMPT_VISIBLE "PROMPT_VISIBLE:"
#define CMD_INFO           "INFO:"
#define CMD_SUCCESS        "SUCCESS"
#define CMD_FAILURE        "FAILURE:"

/* Maximum line length for the pipe protocol. */
#define MAX_LINE 4096

/* Maximum number of PAM module arguments. */
#define MAX_ARGS 32

/* Configuration parsed from PAM module arguments. */
struct pam_authelia_config {
	const char *binary;              /* Path to authelia-pam binary.        */
	const char *url;                 /* Authelia server URL.                */
	const char *auth_level;          /* 1FA, 2FA, or 1FA+2FA (auth-level). */
	const char *cookie_name;         /* Session cookie name (cookie-name).  */
	const char *ca_cert;             /* Custom CA certificate path.         */
	const char *method_priority;     /* Comma-separated ordered method list.*/
	const char *oauth2_client_id;    /* OAuth2 client ID for device flow.   */
	const char *oauth2_client_secret;/* OAuth2 client secret.               */
	const char *oauth2_scope;        /* OAuth2 scopes to request.           */
	int         timeout;             /* Timeout in seconds for child proc.  */
	int         debug;               /* Enable debug logging.               */
};

#endif /* PAM_AUTHELIA_H */
