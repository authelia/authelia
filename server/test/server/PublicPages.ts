
import Server from "../../src/lib/Server";
import BluebirdPromise = require("bluebird");
import speakeasy = require("speakeasy");
import Request = require("request");
import nedb = require("nedb");
import { GlobalDependencies } from "../../types/Dependencies";
import { UserConfiguration } from "../../src/lib/configuration/Configuration";
import { TOTPSecret } from "../../types/TOTPSecret";
import U2FMock = require("./../mocks/u2f");
import Endpoints = require("../../../shared/api");
import Requests = require("../requests");
import Assert = require("assert");
import Sinon = require("sinon");
import Winston = require("winston");
import MockDate = require("mockdate");
import ExpressSession = require("express-session");
import ldapjs = require("ldapjs");

const requestp = BluebirdPromise.promisifyAll(Request) as typeof Request;

const PORT = 8090;
const BASE_URL = "http://localhost:" + PORT;
const requests = Requests(PORT);

describe("Public pages of the server must be accessible without session", function () {
  let server: Server;
  let transporter: object;
  let u2f: U2FMock.U2FMock;

  beforeEach(function () {
    const config: UserConfiguration = {
      port: PORT,
      ldap: {
        url: "ldap://127.0.0.1:389",
        base_dn: "ou=users,dc=example,dc=com",
        user: "cn=admin,dc=example,dc=com",
        password: "password",
      },
      session: {
        secret: "session_secret",
        expiration: 50000,
      },
      storage: {
        local: {
          in_memory: true
        }
      },
      regulation: {
        max_retries: 3,
        ban_time: 5 * 60,
        find_time: 5 * 60
      },
      notifier: {
        gmail: {
          username: "user@example.com",
          password: "password",
          sender: "admin@example.com"
        }
      }
    };

    const ldap_client = {
      bind: Sinon.stub(),
      search: Sinon.stub(),
      modify: Sinon.stub(),
      on: Sinon.spy()
    };
    const ldap = {
      Change: Sinon.spy(),
      createClient: Sinon.spy(function () {
        return ldap_client;
      })
    };

    u2f = U2FMock.U2FMock();

    transporter = {
      sendMail: Sinon.stub().yields()
    };

    const nodemailer = {
      createTransport: Sinon.spy(function () {
        return transporter;
      })
    };

    const ldap_document = {
      object: {
        mail: "test_ok@example.com",
      }
    };

    const search_res = {
      on: Sinon.spy(function (event: string, fn: (s: any) => void) {
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

    const deps: GlobalDependencies = {
      u2f: u2f as any,
      nedb: nedb,
      ldapjs: ldap,
      session: ExpressSession,
      winston: Winston,
      speakeasy: speakeasy,
      ConnectRedis: Sinon.spy()
    };

    server = new Server(deps);
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


  function test_reset_password_form() {
    it("should serve the reset password form page", function (done) {
      requestp.getAsync(BASE_URL + Endpoints.RESET_PASSWORD_REQUEST_GET)
        .then(function (response: Request.RequestResponse) {
          Assert.equal(response.statusCode, 200);
          done();
        });
    });
  }

  function test_login() {
    it("should serve the login page", function (done) {
      requestp.getAsync(BASE_URL + Endpoints.FIRST_FACTOR_GET)
        .then(function (response: Request.RequestResponse) {
          Assert.equal(response.statusCode, 200);
          done();
        });
    });
  }

  function test_logout() {
    it("should logout and redirect to /", function (done) {
      requestp.getAsync(BASE_URL + Endpoints.LOGOUT_GET)
        .then(function (response: any) {
          Assert.equal(response.req.path, "/");
          done();
        });
    });
  }
});

