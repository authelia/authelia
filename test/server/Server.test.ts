
import Server from "../../src/server/lib/Server";
import LdapClient = require("../../src/server/lib/LdapClient");

import BluebirdPromise = require("bluebird");
import speakeasy = require("speakeasy");
import request = require("request");
import nedb = require("nedb");
import { TOTPSecret } from "../../src/types/TOTPSecret";
import U2FMock = require("./mocks/u2f");
import Endpoints = require("../../src/server/endpoints");


const requestp = BluebirdPromise.promisifyAll(request) as typeof request;
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
  let u2f: U2FMock.U2FMock;

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

    u2f = U2FMock.U2FMock();

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

  describe("test GET " + Endpoints.FIRST_FACTOR_GET, function () {
    test_login();
  });

  describe("test GET " + Endpoints.LOGOUT_GET, function () {
    test_logout();
  });

  describe("test GET" + Endpoints.RESET_PASSWORD_REQUEST_GET, function () {
    test_reset_password_form();
  });

  describe("Second factor endpoints must be protected if first factor is not validated", function () {
    function should_post_and_reply_with(url: string, status_code: number): BluebirdPromise<void> {
      return requestp.postAsync(url).then(function (response: request.RequestResponse) {
        assert.equal(response.statusCode, status_code);
        return BluebirdPromise.resolve();
      });
    }

    function should_get_and_reply_with(url: string, status_code: number): BluebirdPromise<void> {
      return requestp.getAsync(url).then(function (response: request.RequestResponse) {
        assert.equal(response.statusCode, status_code);
        return BluebirdPromise.resolve();
      });
    }

    function should_post_and_reply_with_401(url: string): BluebirdPromise<void> {
      return should_post_and_reply_with(url, 401);
    }
    function should_get_and_reply_with_401(url: string): BluebirdPromise<void> {
      return should_get_and_reply_with(url, 401);
    }

    it("should block " + Endpoints.SECOND_FACTOR_GET, function () {
      return should_get_and_reply_with_401(BASE_URL + Endpoints.SECOND_FACTOR_GET);
    });

    it("should block " + Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_GET, function () {
      return should_get_and_reply_with_401(BASE_URL + Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_GET);
    });

    it("should block " + Endpoints.SECOND_FACTOR_U2F_IDENTITY_FINISH_GET, function () {
      return should_get_and_reply_with_401(BASE_URL + Endpoints.SECOND_FACTOR_U2F_IDENTITY_FINISH_GET + "?identity_token=dummy");
    });

    it("should block " + Endpoints.SECOND_FACTOR_U2F_REGISTER_REQUEST_GET, function () {
      return should_get_and_reply_with_401(BASE_URL + Endpoints.SECOND_FACTOR_U2F_REGISTER_REQUEST_GET);
    });

    it("should block " + Endpoints.SECOND_FACTOR_U2F_REGISTER_POST, function () {
      return should_post_and_reply_with_401(BASE_URL + Endpoints.SECOND_FACTOR_U2F_REGISTER_POST);
    });

    it("should block " + Endpoints.SECOND_FACTOR_U2F_SIGN_REQUEST_GET, function () {
      return should_get_and_reply_with_401(BASE_URL + Endpoints.SECOND_FACTOR_U2F_SIGN_REQUEST_GET);
    });

    it("should block " + Endpoints.SECOND_FACTOR_U2F_SIGN_POST, function () {
      return should_post_and_reply_with_401(BASE_URL + Endpoints.SECOND_FACTOR_U2F_SIGN_POST);
    });

    it("should block " + Endpoints.SECOND_FACTOR_TOTP_POST, function () {
      return should_post_and_reply_with_401(BASE_URL + Endpoints.SECOND_FACTOR_TOTP_POST);
    });
  });

  describe("test authentication and verification", function () {
    test_authentication();
    test_reset_password();
    test_regulation();
  });

  function test_reset_password_form() {
    it("should serve the reset password form page", function (done) {
      requestp.getAsync(BASE_URL + Endpoints.RESET_PASSWORD_REQUEST_GET)
        .then(function (response: request.RequestResponse) {
          assert.equal(response.statusCode, 200);
          done();
        });
    });
  }

  function test_login() {
    it("should serve the login page", function (done) {
      requestp.getAsync(BASE_URL + Endpoints.FIRST_FACTOR_GET)
        .then(function (response: request.RequestResponse) {
          assert.equal(response.statusCode, 200);
          done();
        });
    });
  }

  function test_logout() {
    it("should logout and redirect to /", function (done) {
      requestp.getAsync(BASE_URL + Endpoints.LOGOUT_GET)
        .then(function (response: any) {
          assert.equal(response.req.path, "/");
          done();
        });
    });
  }

  function test_authentication() {
    it("should return status code 401 when user is not authenticated", function () {
      return requestp.getAsync({ url: BASE_URL + Endpoints.VERIFY_GET })
        .then(function (response: request.RequestResponse) {
          assert.equal(response.statusCode, 401);
          return BluebirdPromise.resolve();
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
          assert.equal(res.statusCode, 302, "first factor failed");
          return requests.register_totp(j, transporter);
        })
        .then(function (base32_secret: string) {
          const real_token = speakeasy.totp({
            secret: base32_secret,
            encoding: "base32"
          });
          return requests.totp(j, real_token);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 200, "second factor failed");
          return requests.verify(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "verify failed");
          return BluebirdPromise.resolve();
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
          return BluebirdPromise.resolve();
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
      u2f.request.returns(BluebirdPromise.resolve(sign_request));
      u2f.checkRegistration.returns(BluebirdPromise.resolve(sign_status));
      u2f.checkSignature.returns(BluebirdPromise.resolve(registration_status));

      const j = requestp.jar();
      return requests.login(j)
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 200, "get login page failed");
          return requests.first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          // console.log(res);
          assert.equal(res.headers.location, Endpoints.SECOND_FACTOR_GET);
          assert.equal(res.statusCode, 302, "first factor failed");
          return requests.u2f_registration(j, transporter);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 200, "second factor, finish register failed");
          return requests.u2f_authentication(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 200, "second factor, finish sign failed");
          return requests.verify(j);
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "verify failed");
          return BluebirdPromise.resolve();
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
          assert.equal(res.headers.location, Endpoints.SECOND_FACTOR_GET);
          assert.equal(res.statusCode, 302, "first factor failed");
          return requests.reset_password(j, transporter, "user", "new-password");
        })
        .then(function (res: request.RequestResponse) {
          assert.equal(res.statusCode, 204, "second factor, finish register failed");
          return BluebirdPromise.resolve();
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
          return BluebirdPromise.resolve();
        });
    });
  }
});

