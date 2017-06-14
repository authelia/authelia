
import * as BluebirdPromise from "bluebird";
import * as request from "request";

import Server from "../../src/server/lib/Server";
import { UserConfiguration } from "../../src/types/Configuration";
import { GlobalDependencies } from "../../src/types/Dependencies";
import * as tmp from "tmp";
import U2FMock = require("./mocks/u2f");


const requestp = BluebirdPromise.promisifyAll(request) as request.Request;
const assert = require("assert");
const speakeasy = require("speakeasy");
const sinon = require("sinon");
const nedb = require("nedb");
const session = require("express-session");
const winston = require("winston");

const PORT = 8050;
const requests = require("./requests")(PORT);

describe("test data persistence", function () {
  let u2f: U2FMock.U2FMock;
  let tmpDir: tmp.SynchrounousResult;
  const ldap_client = {
    bind: sinon.stub(),
    search: sinon.stub(),
    on: sinon.spy()
  };
  const ldap = {
    createClient: sinon.spy(function () {
      return ldap_client;
    })
  };

  let config: UserConfiguration;

  before(function () {
    u2f = U2FMock.U2FMock();

    const search_doc = {
      object: {
        mail: "test_ok@example.com"
      }
    };

    const search_res = {
      on: sinon.spy(function (event: string, fn: (s: object) => void) {
        if (event != "error") fn(search_doc);
      })
    };

    ldap_client.bind.withArgs("cn=test_ok,ou=users,dc=example,dc=com",
      "password").yields(undefined);
    ldap_client.bind.withArgs("cn=test_nok,ou=users,dc=example,dc=com",
      "password").yields("error");
    ldap_client.search.yields(undefined, search_res);

    tmpDir = tmp.dirSync({ unsafeCleanup: true });
    config = {
      port: PORT,
      ldap: {
        url: "ldap://127.0.0.1:389",
        base_dn: "ou=users,dc=example,dc=com",
        user: "user",
        password: "password"
      },
      session: {
        secret: "session_secret",
        expiration: 50000,
      },
      store_directory: tmpDir.name,
      notifier: {
        gmail: {
          username: "user@example.com",
          password: "password"
        }
      }
    };
  });

  after(function () {
    tmpDir.removeCallback();
  });

  it("should save a u2f meta and reload it after a restart of the server", function () {
    let server: Server;
    const sign_request = {};
    const sign_status = {};
    const registration_status = {};
    u2f.request.returns(sign_request);
    u2f.checkRegistration.returns(sign_status);
    u2f.checkSignature.returns(registration_status);

    const nodemailer = {
      createTransport: sinon.spy(function () {
        return transporter;
      })
    };
    const transporter = {
      sendMail: sinon.stub().yields()
    };

    const deps = {
      u2f: u2f,
      nedb: nedb,
      nodemailer: nodemailer,
      session: session,
      winston: winston,
      ldapjs: ldap,
      speakeasy: speakeasy
    } as GlobalDependencies;

    const j1 = request.jar();
    const j2 = request.jar();

    return start_server(config, deps)
      .then(function (s) {
        server = s;
        return requests.login(j1);
      })
      .then(function (res) {
        return requests.first_factor(j1);
      })
      .then(function () {
        return requests.u2f_registration(j1, transporter);
      })
      .then(function () {
        return requests.u2f_authentication(j1);
      })
      .then(function () {
        return stop_server(server);
      })
      .then(function () {
        return start_server(config, deps);
      })
      .then(function (s) {
        server = s;
        return requests.login(j2);
      })
      .then(function () {
        return requests.first_factor(j2);
      })
      .then(function () {
        return requests.u2f_authentication(j2);
      })
      .then(function (res) {
        assert.equal(200, res.statusCode);
        server.stop();
        return BluebirdPromise.resolve();
      })
      .catch(function (err) {
        console.error(err);
        return BluebirdPromise.reject(err);
      });
  });

  function start_server(config: UserConfiguration, deps: GlobalDependencies): BluebirdPromise<Server> {
    return new BluebirdPromise<Server>(function (resolve, reject) {
      const s = new Server();
      s.start(config, deps);
      resolve(s);
    });
  }

  function stop_server(s: Server) {
    return new BluebirdPromise(function (resolve, reject) {
      s.stop();
      resolve();
    });
  }
});
