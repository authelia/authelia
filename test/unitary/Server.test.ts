
import Server from "../../src/lib/Server";
import LdapClient = require("../../src/lib/LdapClient");

import Promise = require("bluebird");
import speakeasy = require("speakeasy");
import request = require("request");
import nedb = require("nedb");
import { TOTPSecret } from "../../src/types/TOTPSecret";


const requestp = Promise.promisifyAll(request) as request.RequestAsync;
const assert = require("assert");
const sinon = require("sinon");
const MockDate = require("mockdate");
const session = require("express-session");
const winston = require("winston");
const ldapjs = require("ldapjs");

const PORT = 8090;
const BASE_URL = "http://localhost:" + PORT;
const requests = require("./requests")(PORT);

describe("test the server", function () {
  let server: Server;
  let transporter: object;
  let u2f: any;

  beforeEach(function () {
    const config = {
      port: PORT,
      ldap: {
        url: "ldap://127.0.0.1:389",
        base_dn: "ou=users,dc=example,dc=com",
        user_name_attribute: "cn",
        user: "cn=admin,dc=example,dc=com",
        password: "password",
      },
      session: {
        secret: "session_secret",
        expiration: 50000,
      },
      store_in_memory: true,
      notifier: {
        gmail: {
          username: "user@example.com",
          password: "password"
        }
      }
    };

    const ldap_client = {
      bind: sinon.stub(),
      search: sinon.stub(),
      modify: sinon.stub(),
      on: sinon.spy()
    };
    const ldap = {
      Change: sinon.spy(),
      createClient: sinon.spy(function () {
        return ldap_client;
      })
    };

    u2f = {
      startRegistration: sinon.stub(),
      finishRegistration: sinon.stub(),
      startAuthentication: sinon.stub(),
      finishAuthentication: sinon.stub()
    };

    transporter = {
      sendMail: sinon.stub().yields()
    };

    const nodemailer = {
      createTransport: sinon.spy(function () {
        return transporter;
      })
    };

    const ldap_document = {
      object: {
        mail: "test_ok@example.com",
      }
    };

    const search_res = {
      on: sinon.spy(function (event: string, fn: (s: any) => void) {
        if (event != "error") fn(ldap_document);
      })
    };

    ldap_client.bind.withArgs("cn=test_ok,ou=users,dc=example,dc=com",
      "password").yields(undefined);
    ldap_client.bind.withArgs("cn=admin,dc=example,dc=com",
      "password").yields(undefined);

    ldap_client.bind.withArgs("cn=test_nok,ou=users,dc=example,dc=com",
      "password").yields("error");

    ldap_client.modify.yields(undefined);
    ldap_client.search.yields(undefined, search_res);

    const deps = {
      u2f: u2f,
      nedb: nedb,
      nodemailer: nodemailer,
      ldapjs: ldap,
      session: session,
      winston: winston,
      speakeasy: speakeasy
    };

    server = new Server();
    return server.start(config, deps);
  });

  afterEach(function () {
    server.stop();
  });

  describe("test GET /login", function () {
    test_login();
  });

  describe("test GET /logout", function () {
    test_logout();
  });

  describe("test GET /reset-password-form", function () {
    test_reset_password_form();
  });

  describe("test endpoints locks", function () {
    function should_post_and_reply_with(url: string, status_code: number) {
      return requestp.postAsync(url).then(function (response: request.RequestResponse) {
        assert.equal(response.statusCode, status_code);
        return Promise.resolve();
      });
    }

    function should_get_and_reply_with(url: string, status_code: number) {
      return requestp.getAsync(url).then(function (response: request.RequestResponse) {
        assert.equal(response.statusCode, status_code);
        return Promise.resolve();
      });
    }

    function should_post_and_reply_with_403(url: string) {
      return should_post_and_reply_with(url, 403);
    }
    function should_get_and_reply_with_403(url: string) {
      return should_get_and_reply_with(url, 403);
    }

    function should_post_and_reply_with_401(url: string) {
      return should_post_and_reply_with(url, 401);
    }
    function should_get_and_reply_with_401(url: string) {
      return should_get_and_reply_with(url, 401);
    }

    function should_get_and_post_reply_with_403(url: string) {
      const p1 = should_post_and_reply_with_403(url);
      const p2 = should_get_and_reply_with_403(url);
      return Promise.all([p1, p2]);
    }

    it("should block /new-password", function () {
      return should_post_and_reply_with_403(BASE_URL + "/new-password");
    });

    it("should block /u2f-register", function () {
      return should_get_and_post_reply_with_403(BASE_URL + "/u2f-register");
    });

    it("should block /reset-password", function () {
      return should_get_and_post_reply_with_403(BASE_URL + "/reset-password");
    });

    it("should block /2ndfactor/u2f/register_request", function () {
      return should_get_and_reply_with_403(BASE_URL + "/2ndfactor/u2f/register_request");
    });

    it("should block /2ndfactor/u2f/register", function () {
      return should_post_and_reply_with_403(BASE_URL + "/2ndfactor/u2f/register");
    });

    it("should block /2ndfactor/u2f/sign_request", function () {
      return should_get_and_reply_with_403(BASE_URL + "/2ndfactor/u2f/sign_request");
    });

    it("should block /2ndfactor/u2f/sign", function () {
      return should_post_and_reply_with_403(BASE_URL + "/2ndfactor/u2f/sign");
    });
  });

  describe("test authentication and verification", function () {
    test_authentication();
    test_reset_password();
    test_regulation();
  });

  function test_reset_password_form() {
    it("should serve the reset password form page", function (done) {
      requestp.getAsync(BASE_URL + "/reset-password-form")
        .then(function (response: request.RequestResponse) {
          assert.equal(response.statusCode, 200);
          done();
        });
    });
  }

  function test_login() {
    it("should serve the login page", function (done) {
      requestp.getAsync(BASE_URL + "/login")
        .then(function (response: request.RequestResponse) {
          assert.equal(response.statusCode, 200);
          done();
        });
    });
  }

  function test_logout() {
    it("should logout and redirect to /", function (done) {
      requestp.getAsync(BASE_URL + "/logout")
        .then(function (response: any) {
          assert.equal(response.req.path, "/");
          done();
        });
    });
  }

  function test_authentication() {
    it("should return status code 401 when user is not authenticated", function () {
      return requestp.getAsync({ url: BASE_URL + "/verify" })
        .then(function (response: request.RequestResponse) {
          assert.equal(response.statusCode, 401);
          return Promise.resolve();
        });
    });

    it("should return status code 204 when user is authenticated using totp", function () {
      const j = requestp.jar();
      return requests.login(j)
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 200, "get login page failed");
          return requests.first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "first factor failed");
          return requests.register_totp(j, transporter);
        })
        .then(function (secret: string) {
          const sec = JSON.parse(secret) as TOTPSecret;
          const real_token = speakeasy.totp({
            secret: sec.base32,
            encoding: "base32"
          });
          return requests.totp(j, real_token);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "second factor failed");
          return requests.verify(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "verify failed");
          return Promise.resolve();
        });
    });

    it("should keep session variables when login page is reloaded", function () {
      const real_token = speakeasy.totp({
        secret: "totp_secret",
        encoding: "base32"
      });
      const j = requestp.jar();
      return requests.login(j)
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 200, "get login page failed");
          return requests.first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "first factor failed");
          return requests.totp(j, real_token);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "second factor failed");
          return requests.login(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 200, "login page loading failed");
          return requests.verify(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "verify failed");
          return Promise.resolve();
        })
        .catch(function (err: Error) {
          console.error(err);
        });
    });

    it("should return status code 204 when user is authenticated using u2f", function () {
      const sign_request = {};
      const sign_status = {};
      const registration_request = {};
      const registration_status = {};
      u2f.startRegistration.returns(Promise.resolve(sign_request));
      u2f.finishRegistration.returns(Promise.resolve(sign_status));
      u2f.startAuthentication.returns(Promise.resolve(registration_request));
      u2f.finishAuthentication.returns(Promise.resolve(registration_status));

      const j = requestp.jar();
      return requests.login(j)
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 200, "get login page failed");
          return requests.first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "first factor failed");
          return requests.u2f_registration(j, transporter);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "second factor, finish register failed");
          return requests.u2f_authentication(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "second factor, finish sign failed");
          return requests.verify(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "verify failed");
          return Promise.resolve();
        });
    });
  }

  function test_reset_password() {
    it("should reset the password", function () {
      const j = requestp.jar();
      return requests.login(j)
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 200, "get login page failed");
          return requests.first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "first factor failed");
          return requests.reset_password(j, transporter, "user", "new-password");
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "second factor, finish register failed");
          return Promise.resolve();
        });
    });
  }

  function test_regulation() {
    it("should regulate authentication", function () {
      const j = requestp.jar();
      MockDate.set("1/2/2017 00:00:00");
      return requests.login(j)
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 200, "get login page failed");
          return requests.failing_first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 401, "first factor failed");
          return requests.failing_first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 401, "first factor failed");
          return requests.failing_first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 401, "first factor failed");
          return requests.failing_first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 403, "first factor failed");
          MockDate.set("1/2/2017 00:30:00");
          return requests.failing_first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 401, "first factor failed");
          return Promise.resolve();
        });
    });
  }
});

