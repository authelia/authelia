
import Assert = require("assert");
import BluebirdPromise = require("bluebird");
import Express = require("express");
import Sinon = require("sinon");
import winston = require("winston");

import VerifyGet = require("./get");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import ExpressMock = require("../../stubs/express.spec");
import { ServerVariables } from "../../ServerVariables";
import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../ServerVariablesMockBuilder.spec";
import { Level } from "../../authentication/Level";
import { Level as AuthorizationLevel } from "../../authorization/Level";

describe("routes/verify/get", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;
  let authSession: AuthenticationSession;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
    req.originalUrl = "/api/xxxx";
    req.query = {
      redirect: "undefined"
    };
    AuthenticationSessionHandler.reset(req as any);
    req.headers["x-original-url"] = "https://secret.example.com/";
    const s = ServerVariablesMockBuilder.build(false);
    mocks = s.mocks;
    vars = s.variables;
    authSession = AuthenticationSessionHandler.get(req as any, vars.logger);
  });

  describe("with session cookie", function () {
    it("should be already authenticated", function () {
      mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      authSession.authentication_level = Level.TWO_FACTOR;
      authSession.userid = "myuser";
      authSession.groups = ["mygroup", "othergroup"];
      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Sinon.assert.calledWithExactly(res.setHeader, "Remote-User", "myuser");
          Sinon.assert.calledWithExactly(res.setHeader, "Remote-Groups", "mygroup,othergroup");
          Assert.equal(204, res.status.getCall(0).args[0]);
        });
    });

    function test_session(_authSession: AuthenticationSession, status_code: number) {
      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Assert.equal(status_code, res.status.getCall(0).args[0]);
        });
    }

    function test_non_authenticated_401(authSession: AuthenticationSession) {
      return test_session(authSession, 401);
    }

    function test_unauthorized_403(authSession: AuthenticationSession) {
      return test_session(authSession, 403);
    }

    function test_authorized(authSession: AuthenticationSession) {
      return test_session(authSession, 204);
    }

    describe("given user tries to access a 2-factor endpoint", function () {
      before(function () {
        mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      });

      describe("given different cases of session", function () {
        it("should not be authenticated when second factor is missing", function () {
          return test_non_authenticated_401({
            keep_me_logged_in: false,
            userid: "user",
            authentication_level: Level.ONE_FACTOR,
            email: undefined,
            groups: [],
            last_activity_datetime: new Date().getTime()
          });
        });

        it("should not be authenticated when userid is missing", function () {
          return test_non_authenticated_401({
            keep_me_logged_in: false,
            userid: undefined,
            authentication_level: Level.TWO_FACTOR,
            email: undefined,
            groups: [],
            last_activity_datetime: new Date().getTime()
          });
        });

        it("should not be authenticated when level is insufficient", function () {
          return test_non_authenticated_401({
            keep_me_logged_in: false,
            userid: "user",
            authentication_level: Level.NOT_AUTHENTICATED,
            email: undefined,
            groups: [],
            last_activity_datetime: new Date().getTime()
          });
        });

        it("should not be authenticated when session has not be initiated", function () {
          return test_non_authenticated_401(undefined);
        });

        it("should not be authenticated when domain is not allowed for user", function () {
          authSession.authentication_level = Level.TWO_FACTOR;
          authSession.userid = "myuser";
          req.headers["x-original-url"] = "https://test.example.com/";
          mocks.authorizer.authorizationMock.returns(AuthorizationLevel.DENY);

          return test_unauthorized_403({
            keep_me_logged_in: false,
            authentication_level: Level.TWO_FACTOR,
            userid: "user",
            groups: ["group1", "group2"],
            email: undefined,
            last_activity_datetime: new Date().getTime()
          });
        });
      });
    });

    describe("given user tries to access a single factor endpoint", function () {
      beforeEach(function () {
        req.headers["x-original-url"] = "https://redirect.url/";
      });

      it("should be authenticated when first factor is validated", function () {
        mocks.authorizer.authorizationMock.returns(AuthorizationLevel.ONE_FACTOR);
        authSession.authentication_level = Level.ONE_FACTOR;
        authSession.userid = "user1";
        return VerifyGet.default(vars)(req as Express.Request, res as any)
          .then(function () {
            Assert(res.status.calledWith(204));
            Assert(res.send.calledOnce);
          });
      });

      it("should be rejected with 401 when not authenticated", function () {
        mocks.authorizer.authorizationMock.returns(AuthorizationLevel.ONE_FACTOR);
        authSession.authentication_level = Level.NOT_AUTHENTICATED;
        return VerifyGet.default(vars)(req as Express.Request, res as any)
          .then(function () {
            Assert(res.status.calledWith(401));
          });
      });
    });

    describe("inactivity period", function () {
      it("should update last inactivity period on requests on /api/verify", function () {
        mocks.config.session.inactivity = 200000;
        mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
        const currentTime = new Date().getTime() - 1000;
        AuthenticationSessionHandler.reset(req as any);
        authSession.authentication_level = Level.TWO_FACTOR;
        authSession.userid = "myuser";
        authSession.groups = ["mygroup", "othergroup"];
        authSession.last_activity_datetime = currentTime;
        return VerifyGet.default(vars)(req as Express.Request, res as any)
          .then(function () {
            return AuthenticationSessionHandler.get(req as any, vars.logger);
          })
          .then(function (authSession) {
            Assert(authSession.last_activity_datetime > currentTime);
          });
      });

      it("should reset session when max inactivity period has been reached", function () {
        mocks.config.session.inactivity = 1;
        mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
        const currentTime = new Date().getTime() - 1000;
        AuthenticationSessionHandler.reset(req as any);
        authSession.authentication_level = Level.TWO_FACTOR;
        authSession.userid = "myuser";
        authSession.groups = ["mygroup", "othergroup"];
        authSession.last_activity_datetime = currentTime;
        return VerifyGet.default(vars)(req as Express.Request, res as any)
          .then(function () {
            return AuthenticationSessionHandler.get(req as any, vars.logger);
          })
          .then(function (authSession) {
            Assert.equal(authSession.authentication_level, Level.NOT_AUTHENTICATED);
            Assert.equal(authSession.userid, undefined);
          });
      });
    });
  });

  describe("response type 401 | 302", function() {
    it("should return error code 401", function() {
      mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      mocks.config.access_control.default_policy = "one_factor";
      mocks.usersDatabase.checkUserPasswordStub.rejects(new Error(
        "Invalid credentials"));
      req.headers["proxy-authorization"] = "Basic am9objpwYXNzd29yZA==";

      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Assert(res.status.calledWithExactly(401));
        });
    });

    it("should redirect to provided redirection url", function() {
      const REDIRECT_URL = "http://redirection_url.com";
      mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      mocks.config.access_control.default_policy = "one_factor";
      mocks.usersDatabase.checkUserPasswordStub.rejects(new Error(
        "Invalid credentials"));
      req.headers["proxy-authorization"] = "Basic am9objpwYXNzd29yZA==";
      req.query["rd"] = REDIRECT_URL;

      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Assert(res.redirect.calledWithExactly(REDIRECT_URL));
        });
    });
  });

  describe("with basic auth", function () {
    it("should authenticate correctly", function () {
      mocks.authorizer.authorizationMock.returns(AuthorizationLevel.ONE_FACTOR);
      mocks.config.access_control.default_policy = "one_factor";
      mocks.usersDatabase.checkUserPasswordStub.returns({
        groups: ["mygroup", "othergroup"],
      });
      req.headers["proxy-authorization"] = "Basic am9objpwYXNzd29yZA==";

      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Sinon.assert.calledWithExactly(res.setHeader, "Remote-User", "john");
          Sinon.assert.calledWithExactly(res.setHeader, "Remote-Groups", "mygroup,othergroup");
          Assert.equal(204, res.status.getCall(0).args[0]);
        });
    });

    it("should fail when endpoint is protected by two factors", function () {
      mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      mocks.config.access_control.default_policy = "one_factor";
      mocks.config.access_control.rules = [{
        domain: "secret.example.com",
        policy: "two_factor"
      }];
      mocks.usersDatabase.checkUserPasswordStub.resolves({
        groups: ["mygroup", "othergroup"],
      });
      req.headers["proxy-authorization"] = "Basic am9objpwYXNzd29yZA==";

      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Assert(res.status.calledWithExactly(401));
        });
    });

    it("should fail when base64 token is not valid", function () {
      mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      mocks.config.access_control.default_policy = "one_factor";
      mocks.usersDatabase.checkUserPasswordStub.resolves({
        groups: ["mygroup", "othergroup"],
      });
      req.headers["proxy-authorization"] = "Basic i_m*not_a_base64*token";

      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Assert(res.status.calledWithExactly(401));
        });
    });

    it("should fail when base64 token has not format user:psswd", function () {
      mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      mocks.config.access_control.default_policy = "one_factor";
      mocks.usersDatabase.checkUserPasswordStub.resolves({
        groups: ["mygroup", "othergroup"],
      });
      req.headers["proxy-authorization"] = "Basic am9objpwYXNzOmJhZA==";

      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Assert(res.status.calledWithExactly(401));
        });
    });

    it("should fail when bad user password is provided", function () {
      mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      mocks.config.access_control.default_policy = "one_factor";
      mocks.usersDatabase.checkUserPasswordStub.rejects(new Error(
        "Invalid credentials"));
      req.headers["proxy-authorization"] = "Basic am9objpwYXNzd29yZA==";

      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Assert(res.status.calledWithExactly(401));
        });
    });

    it("should fail when resource is restricted", function () {
      mocks.authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      mocks.config.access_control.default_policy = "one_factor";
      mocks.usersDatabase.checkUserPasswordStub.resolves({
        groups: ["mygroup", "othergroup"],
      });
      req.headers["proxy-authorization"] = "Basic am9objpwYXNzd29yZA==";

      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Assert(res.status.calledWithExactly(401));
        });
    });
  });
});

