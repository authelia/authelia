import Server from "../../../../src/server/lib/Server";
import { LdapjsClientMock } from "./../mocks/ldapjs";

import BluebirdPromise = require("bluebird");
import speakeasy = require("speakeasy");
import request = require("request");
import nedb = require("nedb");
import { GlobalDependencies } from "../../../../src/types/Dependencies";
import { TOTPSecret } from "../../../../src/types/TOTPSecret";
import U2FMock = require("./../mocks/u2f");
import Endpoints = require("../../../../src/server/endpoints");
import Requests = require("../requests");
import Assert = require("assert");
import Sinon = require("sinon");
import Winston = require("winston");
import MockDate = require("mockdate");
import ExpressSession = require("express-session");
import ldapjs = require("ldapjs");

const requestp = BluebirdPromise.promisifyAll(request) as typeof request;

const PORT = 8090;
const BASE_URL = "http://localhost:" + PORT;
const requests = Requests(PORT);

describe("test the server", function () {
  let server: Server;
  let transporter: any;
  let u2f: U2FMock.U2FMock;

  beforeEach(function () {
    const config = {
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
      notifier: {
        gmail: {
          username: "user@example.com",
          password: "password"
        }
      }
    };

    const ldapClient = LdapjsClientMock();
    const ldap = {
      Change: Sinon.spy(),
      createClient: Sinon.spy(function () {
        return ldapClient;
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

    const ldapDocument = {
      object: {
        mail: "test_ok@example.com",
      }
    };

    const search_res = {
      on: Sinon.spy(function (event: string, fn: (s: any) => void) {
        if (event != "error") fn(ldapDocument);
      })
    };

    ldapClient.bind.withArgs("cn=test_ok,ou=users,dc=example,dc=com",
      "password").yields();
    ldapClient.bind.withArgs("cn=admin,dc=example,dc=com",
      "password").yields();

    ldapClient.bind.withArgs("cn=test_nok,ou=users,dc=example,dc=com",
      "password").yields("Bad credentials");

    ldapClient.unbind.yields();
    ldapClient.modify.yields();
    ldapClient.search.yields(undefined, search_res);

    const deps: GlobalDependencies = {
      u2f: u2f,
      nedb: nedb,
      nodemailer: nodemailer,
      ldapjs: ldap,
      session: ExpressSession,
      winston: Winston,
      speakeasy: speakeasy,
      ConnectRedis: Sinon.spy(),
      dovehash: {
        encode: Sinon.stub().returns("abc")
      }
    };

    server = new Server();
    return server.start(config, deps);
  });

  afterEach(function () {
    server.stop();
  });

  describe("test authentication and verification", function () {
    test_authentication();
    test_regulation();
  });

  function test_authentication() {
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
          Assert.equal(res.statusCode, 200, "get login page failed");
          return requests.first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          // console.log(res);
          Assert.equal(res.headers.location, Endpoints.SECOND_FACTOR_GET);
          Assert.equal(res.statusCode, 302, "first factor failed");
          return requests.u2f_registration(j, transporter);
        })
        .then(function (res: request.RequestResponse) {
          Assert.equal(res.statusCode, 200, "second factor, finish register failed");
          return requests.u2f_authentication(j);
        })
        .then(function (res: request.RequestResponse) {
          Assert.equal(res.statusCode, 200, "second factor, finish sign failed");
          return requests.verify(j);
        })
        .then(function (res: request.RequestResponse) {
          Assert.equal(res.statusCode, 204, "verify failed");
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
          Assert.equal(res.statusCode, 200, "get login page failed");
          return requests.failing_first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          Assert.equal(res.statusCode, 401, "first factor failed");
          return requests.failing_first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          Assert.equal(res.statusCode, 401, "first factor failed");
          return requests.failing_first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          Assert.equal(res.statusCode, 401, "first factor failed");
          return requests.failing_first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          Assert.equal(res.statusCode, 403, "first factor failed");
          MockDate.set("1/2/2017 00:30:00");
          return requests.failing_first_factor(j);
        })
        .then(function (res: request.RequestResponse) {
          Assert.equal(res.statusCode, 401, "first factor failed");
          return BluebirdPromise.resolve();
        });
    });
  }
});

