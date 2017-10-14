
import Server from "../../src/lib/Server";
import BluebirdPromise = require("bluebird");
import speakeasy = require("speakeasy");
import request = require("request");
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

const requestp = BluebirdPromise.promisifyAll(request) as typeof request;

const PORT = 8090;
const BASE_URL = "http://localhost:" + PORT;
const requests = Requests(PORT);

describe("Private pages of the server must not be accessible without session", function () {
  let server: Server;
  let transporter: any;
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
      regulation: {
        max_retries: 3,
        ban_time: 5 * 60,
        find_time: 5 * 60
      },
      storage: {
        local: {
          in_memory: true
        }
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
      ConnectRedis: Sinon.spy(),
      dovehash: Sinon.spy() as any
    };

    server = new Server(deps);
    return server.start(config, deps);
  });

  afterEach(function () {
    server.stop();
  });

  describe("Second factor endpoints must be protected if first factor is not validated", function () {
    function should_post_and_reply_with_401(url: string): BluebirdPromise<void> {
      return requestp.postAsync(url).then(function (response: request.RequestResponse) {
        Assert.equal(response.statusCode, 401);
        return BluebirdPromise.resolve();
      });
    }

    function should_get_and_reply_with_401(url: string): BluebirdPromise<void> {
      return requestp.getAsync(url).then(function (response: request.RequestResponse) {
        Assert.equal(response.statusCode, 401);
        return BluebirdPromise.resolve();
      });
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
});

