
import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");
import Winston = require("winston");

import FirstFactorPost = require("../../../src/lib/routes/firstfactor/post");
import exceptions = require("../../../src/lib/Exceptions");
import AuthenticationSession = require("../../../src/lib/AuthenticationSession");
import Endpoints = require("../../../../shared/api");

import AuthenticationRegulatorMock = require("../../mocks/AuthenticationRegulator");
import { AccessControllerStub } from "../../mocks/AccessControllerStub";
import ExpressMock = require("../../mocks/express");
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../mocks/ServerVariablesMockBuilder";
import { ServerVariables } from "../../../src/lib/ServerVariables";

describe("test the first factor validation route", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let emails: string[];
  let groups: string[];
  let vars: ServerVariables;
  let mocks: ServerVariablesMock;

  beforeEach(function () {
    emails = ["test_ok@example.com"];
    groups = ["group1", "group2" ];
    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    mocks.accessController.isAccessAllowedMock.returns(true);
    mocks.regulator.regulateStub.returns(BluebirdPromise.resolve());
    mocks.regulator.markStub.returns(BluebirdPromise.resolve());

    req = {
      app: {
        get: Sinon.stub().returns({ logger: Winston })
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
    res = ExpressMock.ResponseMock();
  });

  it("should reply with 204 if success", function () {
    mocks.ldapAuthenticator.authenticateStub.withArgs("username", "password")
      .returns(BluebirdPromise.resolve({
        emails: emails,
        groups: groups
      }));
    let authSession: AuthenticationSession.AuthenticationSession;
    return AuthenticationSession.get(req as any)
      .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
        authSession = _authSession;
        return FirstFactorPost.default(vars)(req as any, res as any);
      })
      .then(function () {
        Assert.equal("username", authSession.userid);
        Assert(res.send.calledOnce);
      });
  });

  it("should retrieve email from LDAP", function () {
    mocks.ldapAuthenticator.authenticateStub.withArgs("username", "password")
      .returns(BluebirdPromise.resolve([{ mail: ["test@example.com"] }]));
    return FirstFactorPost.default(vars)(req as any, res as any);
  });

  it("should set first email address as user session variable", function () {
    const emails = ["test_ok@example.com"];
    let authSession: AuthenticationSession.AuthenticationSession;
    mocks.ldapAuthenticator.authenticateStub.withArgs("username", "password")
      .returns(BluebirdPromise.resolve({
        emails: emails,
        groups: groups
      }));

    return AuthenticationSession.get(req as any)
      .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
        authSession = _authSession;
        return FirstFactorPost.default(vars)(req as any, res as any);
      })
      .then(function () {
        Assert.equal("test_ok@example.com", authSession.email);
      });
  });

  it("should return error message when LDAP authenticator throws", function () {
    mocks.ldapAuthenticator.authenticateStub.withArgs("username", "password")
      .returns(BluebirdPromise.reject(new exceptions.LdapBindError("Bad credentials")));

    return FirstFactorPost.default(vars)(req as any, res as any)
      .then(function () {
        Assert.equal(res.status.getCall(0).args[0], 200);
        Assert.equal(mocks.regulator.markStub.getCall(0).args[0], "username");
        Assert.deepEqual(res.send.getCall(0).args[0], {
          error: "Operation failed."
        });
      });
  });

  it("should return error message when regulator rejects authentication", function () {
    const err = new exceptions.AuthenticationRegulationError("Authentication regulation...");
    mocks.regulator.regulateStub.returns(BluebirdPromise.reject(err));
    return FirstFactorPost.default(vars)(req as any, res as any)
      .then(function () {
        Assert.equal(res.status.getCall(0).args[0], 200);
        Assert.deepEqual(res.send.getCall(0).args[0], {
          error: "Operation failed."
        });
      });
  });
});


