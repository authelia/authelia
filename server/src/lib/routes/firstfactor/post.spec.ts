import * as Express from 'express';
import BluebirdPromise = require("bluebird");
import Assert = require("assert");
import FirstFactorPost = require("./post");
import exceptions = require("../../Exceptions");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import * as ExpressMock from "../../stubs/express.spec";
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../ServerVariables";
import AuthenticationError from "../../authentication/AuthenticationError";

describe("routes/firstfactor/post", function () {
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;
  let emails: string[];
  let groups: string[];
  let vars: ServerVariables;
  let mocks: ServerVariablesMock;
  let authSession: AuthenticationSession;

  beforeEach(function () {
    emails = ["test_ok@example.com"];
    groups = ["group1", "group2" ];
    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    mocks.authorizer.authorizationMock.returns(true);
    mocks.regulator.regulateStub.returns(BluebirdPromise.resolve());
    mocks.regulator.markStub.returns(BluebirdPromise.resolve());

    req = ExpressMock.RequestMock();
    req.body = {
      username: "username",
      password: "password"
    }
    res = ExpressMock.ResponseMock();
    authSession = AuthenticationSessionHandler.get(req as any, vars.logger);
  });

  it("should reply with 204 if success", function () {
    mocks.usersDatabase.checkUserPasswordStub.withArgs("username", "password")
      .returns(BluebirdPromise.resolve({
        emails: emails,
        groups: groups
      }));
    return FirstFactorPost.default(vars)(req as any, res as any)
      .then(function () {
        Assert.equal("username", authSession.userid);
        Assert(res.send.calledOnce);
      });
  });

  describe("keep me logged in", () => {
    beforeEach(() => {
      mocks.usersDatabase.checkUserPasswordStub.withArgs("username", "password")
        .returns(BluebirdPromise.resolve({
          emails: emails,
          groups: groups
        }));
      req.body.keepMeLoggedIn = true;
      return FirstFactorPost.default(vars)(req as any, res as any);
    });

    it("should set keep_me_logged_in session variable to true", function () {
      Assert.equal(authSession.keep_me_logged_in, true);
    });

    it("should set cookie maxAge to one year", function () {
      Assert.equal(req.session.cookie.maxAge, 365 * 24 * 60 * 60 * 1000);
    });
  });

  it("should retrieve email from LDAP", function () {
    mocks.usersDatabase.checkUserPasswordStub.withArgs("username", "password")
      .returns(BluebirdPromise.resolve([{ mail: ["test@example.com"] }]));
    return FirstFactorPost.default(vars)(req as any, res as any);
  });

  it("should set first email address as user session variable", function () {
    const emails = ["test_ok@example.com"];
    mocks.usersDatabase.checkUserPasswordStub.withArgs("username", "password")
      .returns(BluebirdPromise.resolve({
        emails: emails,
        groups: groups
      }));

      return FirstFactorPost.default(vars)(req as any, res as any)
      .then(function () {
        Assert.equal("test_ok@example.com", authSession.email);
      });
  });

  it("should return error message when LDAP authenticator throws", function () {
    mocks.usersDatabase.checkUserPasswordStub.withArgs("username", "password")
      .returns(BluebirdPromise.reject(new AuthenticationError("Bad credentials")));

    return FirstFactorPost.default(vars)(req as any, res as any)
      .then(function () {
        Assert.equal(res.status.getCall(0).args[0], 200);
        Assert.equal(mocks.regulator.markStub.getCall(0).args[0], "username");
        Assert.deepEqual(res.send.getCall(0).args[0], {
          error: "Authentication failed. Please check your credentials."
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
          error: "Authentication failed. Please check your credentials."
        });
      });
  });
});


