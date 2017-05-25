
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import winston = require("winston");

import FirstFactorPost = require("../../../../src/server/lib/routes/firstfactor/post");
import exceptions = require("../../../../src/server/lib/Exceptions");
import AuthenticationSession = require("../../../../src/server/lib/AuthenticationSession");
import Endpoints = require("../../../../src/server/endpoints");

import AuthenticationRegulatorMock = require("../../mocks/AuthenticationRegulator");
import AccessControllerMock = require("../../mocks/AccessController");
import { LdapClientMock } from "../../mocks/LdapClient";
import ExpressMock = require("../../mocks/express");
import ServerVariablesMock = require("../../mocks/ServerVariablesMock");

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

    req = {
      app: {
      },
      body: {
        username: "username",
        password: "password"
      },
      session: {
      },
      headers: {
        host: "home.example.com"
      }
    };

    AuthenticationSession.reset(req as any);

    const mocks = ServerVariablesMock.mock(req.app);
    mocks.ldap = ldapMock;
    mocks.config = configuration;
    mocks.logger = winston;
    mocks.regulator = regulator;
    mocks.accessController = accessController;

    res = ExpressMock.ResponseMock();
  });

  it("should redirect client to second factor page", function () {
    ldapMock.bind.withArgs("username").returns(BluebirdPromise.resolve());
    ldapMock.get_emails.returns(BluebirdPromise.resolve(emails));
    const authSession = AuthenticationSession.get(req as any);
    return FirstFactorPost.default(req as any, res as any)
      .then(function () {
        assert.equal("username", authSession.userid);
        assert.equal(Endpoints.SECOND_FACTOR_GET, res.redirect.getCall(0).args[0]);
      });
  });

  it("should retrieve email from LDAP", function (done) {
    res.redirect = sinon.spy(function () { done(); });
    ldapMock.bind.returns(BluebirdPromise.resolve());
    ldapMock.get_emails = sinon.stub().withArgs("username").returns(BluebirdPromise.resolve([{ mail: ["test@example.com"] }]));
    FirstFactorPost.default(req as any, res as any);
  });

  it("should set email as session variables", function () {
    const emails = ["test_ok@example.com"];
    const authSession = AuthenticationSession.get(req as any);
    ldapMock.bind.returns(BluebirdPromise.resolve());
    ldapMock.get_emails.returns(BluebirdPromise.resolve(emails));
    return FirstFactorPost.default(req as any, res as any)
      .then(function () {
        assert.equal("test_ok@example.com", authSession.email);
      });
  });

  it("should return status code 401 when LDAP binding throws", function (done) {
    res.send = sinon.spy(function () {
      assert.equal(401, res.status.getCall(0).args[0]);
      assert.equal(regulator.mark.getCall(0).args[0], "username");
      done();
    });
    ldapMock.bind.returns(BluebirdPromise.reject(new exceptions.LdapBindError("Bad credentials")));
    FirstFactorPost.default(req as any, res as any);
  });

  it("should return status code 500 when LDAP search throws", function (done) {
    res.send = sinon.spy(function () {
      assert.equal(500, res.status.getCall(0).args[0]);
      done();
    });
    ldapMock.bind.returns(BluebirdPromise.resolve());
    ldapMock.get_emails.returns(BluebirdPromise.reject(new exceptions.LdapSearchError("error while retrieving emails")));
    FirstFactorPost.default(req as any, res as any);
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
    FirstFactorPost.default(req as any, res as any);
  });
});


