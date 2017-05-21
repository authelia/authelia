
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import winston = require("winston");

import FirstFactor = require("../../../src/lib/routes/FirstFactor");
import exceptions = require("../../../src/lib/Exceptions");
import AuthenticationRegulatorMock = require("../mocks/AuthenticationRegulator");
import AccessControllerMock = require("../mocks/AccessController");
import { LdapClientMock } from "../mocks/LdapClient";
import ExpressMock = require("../mocks/express");

describe("test the first factor validation route", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let emails: string[];
  let groups: string[];
  let configuration;
  let ldapMock: LdapClientMock;
  let regulator: AuthenticationRegulatorMock.AuthenticationRegulatorMock;
  let accessController: AccessControllerMock.AccessControllerMock;

  beforeEach(function () {
    configuration = {
      ldap: {
        base_dn: "ou=users,dc=example,dc=com",
        user_name_attribute: "uid"
      }
    };

    emails = ["test_ok@example.com"];
    groups = ["group1", "group2" ];

    ldapMock = LdapClientMock();

    accessController = AccessControllerMock.AccessControllerMock();
    accessController.isDomainAllowedForUser.returns(true);

    regulator = AuthenticationRegulatorMock.AuthenticationRegulatorMock();
    regulator.regulate.returns(BluebirdPromise.resolve());
    regulator.mark.returns(BluebirdPromise.resolve());

    const app_get = sinon.stub();
    app_get.withArgs("ldap").returns(ldapMock);
    app_get.withArgs("configuration").returns(configuration);
    app_get.withArgs("logger").returns(winston);
    app_get.withArgs("authentication regulator").returns(regulator);
    app_get.withArgs("access controller").returns(accessController);

    req = {
      app: {
        get: app_get
      },
      body: {
        username: "username",
        password: "password"
      },
      session: {
        auth_session: {
          FirstFactor: false,
          second_factor: false
        }
      },
      headers: {
        host: "home.example.com"
      }
    };
    res = ExpressMock.ResponseMock();
  });

  it("should return status code 204 when LDAP binding succeeds", function () {
    return new Promise(function (resolve, reject) {
      res.send = sinon.spy(function () {
        assert.equal("username", req.session.auth_session.userid);
        assert.equal(204, res.status.getCall(0).args[0]);
        resolve();
      });
      ldapMock.bind.withArgs("username").returns(BluebirdPromise.resolve());
      ldapMock.get_emails.returns(BluebirdPromise.resolve(emails));
      FirstFactor(req as any, res as any);
    });
  });

  it("should retrieve email from LDAP", function (done) {
    res.send = sinon.spy(function () { done(); });
    ldapMock.bind.returns(BluebirdPromise.resolve());
    ldapMock.get_emails = sinon.stub().withArgs("username").returns(BluebirdPromise.resolve([{ mail: ["test@example.com"] }]));
    FirstFactor(req as any, res as any);
  });

  it("should set email as session variables", function () {
    return new Promise(function (resolve, reject) {
      res.send = sinon.spy(function () {
        assert.equal("test_ok@example.com", req.session.auth_session.email);
        resolve();
      });
      const emails = ["test_ok@example.com"];
      ldapMock.bind.returns(BluebirdPromise.resolve());
      ldapMock.get_emails.returns(BluebirdPromise.resolve(emails));
      FirstFactor(req as any, res as any);
    });
  });

  it("should return status code 401 when LDAP binding throws", function (done) {
    res.send = sinon.spy(function () {
      assert.equal(401, res.status.getCall(0).args[0]);
      assert.equal(regulator.mark.getCall(0).args[0], "username");
      done();
    });
    ldapMock.bind.returns(BluebirdPromise.reject(new exceptions.LdapBindError("Bad credentials")));
    FirstFactor(req as any, res as any);
  });

  it("should return status code 500 when LDAP search throws", function (done) {
    res.send = sinon.spy(function () {
      assert.equal(500, res.status.getCall(0).args[0]);
      done();
    });
    ldapMock.bind.returns(BluebirdPromise.resolve());
    ldapMock.get_emails.returns(BluebirdPromise.reject(new exceptions.LdapSeachError("error while retrieving emails")));
    FirstFactor(req as any, res as any);
  });

  it("should return status code 403 when regulator rejects authentication", function (done) {
    const err = new exceptions.AuthenticationRegulationError("Authentication regulation...");
    regulator.regulate.returns(BluebirdPromise.reject(err));

    res.send = sinon.spy(function () {
      assert.equal(403, res.status.getCall(0).args[0]);
      done();
    });
    ldapMock.bind.returns(BluebirdPromise.resolve());
    ldapMock.get_emails.returns(BluebirdPromise.resolve());
    FirstFactor(req as any, res as any);
  });
});


