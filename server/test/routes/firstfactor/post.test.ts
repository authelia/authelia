
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");
import winston = require("winston");

import FirstFactorPost = require("../../../src/lib/routes/firstfactor/post");
import exceptions = require("../../../src/lib/Exceptions");
import AuthenticationSession = require("../../../src/lib/AuthenticationSession");
import Endpoints = require("../../../../shared/api");

import AuthenticationRegulatorMock = require("../../mocks/AuthenticationRegulator");
import { AccessControllerStub } from "../../mocks/AccessControllerStub";
import ExpressMock = require("../../mocks/express");
import ServerVariablesMock = require("../../mocks/ServerVariablesMock");
import { ServerVariables } from "../../../src/lib/ServerVariables";

describe("test the first factor validation route", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let emails: string[];
  let groups: string[];
  let configuration;
  let regulator: AuthenticationRegulatorMock.AuthenticationRegulatorMock;
  let accessController: AccessControllerStub;
  let serverVariables: ServerVariables;

  beforeEach(function () {
    configuration = {
      ldap: {
        base_dn: "ou=users,dc=example,dc=com",
        user_name_attribute: "uid"
      }
    };

    emails = ["test_ok@example.com"];
    groups = ["group1", "group2" ];

    accessController = new AccessControllerStub();
    accessController.isAccessAllowedMock.returns(true);

    regulator = AuthenticationRegulatorMock.AuthenticationRegulatorMock();
    regulator.regulate.returns(BluebirdPromise.resolve());
    regulator.mark.returns(BluebirdPromise.resolve());

    req = {
      app: {
        get: sinon.stub().returns({ logger: winston })
      },
      body: {
        username: "username",
        password: "password"
      },
      query: {
        redirect: "http://redirect.url"
      },
      session: {
      },
      headers: {
        host: "home.example.com"
      }
    };

    AuthenticationSession.reset(req as any);

    serverVariables = ServerVariablesMock.mock(req.app);
    serverVariables.ldapAuthenticator = {
      authenticate: sinon.stub()
    } as any;
    serverVariables.config = configuration as any;
    serverVariables.regulator = regulator as any;
    serverVariables.accessController = accessController as any;

    res = ExpressMock.ResponseMock();
  });

  it("should reply with 204 if success", function () {
    (serverVariables.ldapAuthenticator as any).authenticate.withArgs("username", "password")
      .returns(BluebirdPromise.resolve({
        emails: emails,
        groups: groups
      }));
    let authSession: AuthenticationSession.AuthenticationSession;
    return AuthenticationSession.get(req as any)
      .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
        authSession = _authSession;
        return FirstFactorPost.default(req as any, res as any);
      })
      .then(function () {
        Assert.equal("username", authSession.userid);
        Assert(res.send.calledOnce);
      });
  });

  it("should retrieve email from LDAP", function () {
    (serverVariables.ldapAuthenticator as any).authenticate.withArgs("username", "password")
      .returns(BluebirdPromise.resolve([{ mail: ["test@example.com"] }]));
    return FirstFactorPost.default(req as any, res as any);
  });

  it("should set first email address as user session variable", function () {
    const emails = ["test_ok@example.com"];
    let authSession: AuthenticationSession.AuthenticationSession;
    (serverVariables.ldapAuthenticator as any).authenticate.withArgs("username", "password")
      .returns(BluebirdPromise.resolve({
        emails: emails,
        groups: groups
      }));

    return AuthenticationSession.get(req as any)
      .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
        authSession = _authSession;
        return FirstFactorPost.default(req as any, res as any);
      })
      .then(function () {
        Assert.equal("test_ok@example.com", authSession.email);
      });
  });

  it("should return error message when LDAP authenticator throws", function () {
    (serverVariables.ldapAuthenticator as any).authenticate.withArgs("username", "password")
      .returns(BluebirdPromise.reject(new exceptions.LdapBindError("Bad credentials")));

    return FirstFactorPost.default(req as any, res as any)
      .then(function () {
        Assert.equal(res.status.getCall(0).args[0], 200);
        Assert.equal(regulator.mark.getCall(0).args[0], "username");
        Assert.deepEqual(res.send.getCall(0).args[0], {
          error: "Operation failed."
        });
      });
  });

  it("should return error message when regulator rejects authentication", function () {
    const err = new exceptions.AuthenticationRegulationError("Authentication regulation...");
    regulator.regulate.returns(BluebirdPromise.reject(err));
    return FirstFactorPost.default(req as any, res as any)
      .then(function () {
        Assert.equal(res.status.getCall(0).args[0], 200);
        Assert.deepEqual(res.send.getCall(0).args[0], {
          error: "Operation failed."
        });
      });
  });
});


