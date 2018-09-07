import Assert = require("assert");
import Express = require("express");
import Sinon = require("sinon");

import VerifyGet = require("./get");
import ExpressMock = require("../../stubs/express.spec");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import { ServerVariables } from "../../ServerVariables";
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../ServerVariablesMockBuilder.spec";
import { WhitelistValue } from "../../authentication/whitelist/WhitelistHandler";

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
    beforeEach(function () {
      vars.config.authentication_methods.default_method = "two_factor";
    });

    it("should be already authenticated", function () {
      mocks.accessController.isAccessAllowedMock.returns(true);
      authSession.first_factor = true;
      authSession.second_factor = true;
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
        mocks.accessController.isAccessAllowedMock.returns(true);
      });

      describe("given different cases of session", function () {
        it("should not be authenticated when second factor is missing", function () {
          return test_non_authenticated_401({
            userid: "user",
            first_factor: true,
            second_factor: false,
            whitelisted: WhitelistValue.NOT_WHITELISTED,
            email: undefined,
            groups: [],
            last_activity_datetime: new Date().getTime()
          });
        });

        it("should not be authenticated when first factor is missing", function () {
          return test_non_authenticated_401({
            userid: "user",
            first_factor: false,
            second_factor: true,
            whitelisted: WhitelistValue.NOT_WHITELISTED,
            email: undefined,
            groups: [],
            last_activity_datetime: new Date().getTime()
          });
        });

        it("should not be authenticated when userid is missing", function () {
          return test_non_authenticated_401({
            userid: undefined,
            first_factor: true,
            second_factor: false,
            whitelisted: WhitelistValue.NOT_WHITELISTED,
            email: undefined,
            groups: [],
            last_activity_datetime: new Date().getTime()
          });
        });

        it("should not be authenticated when first and second factor are missing", function () {
          return test_non_authenticated_401({
            userid: "user",
            first_factor: false,
            second_factor: false,
            whitelisted: WhitelistValue.NOT_WHITELISTED,
            email: undefined,
            groups: [],
            last_activity_datetime: new Date().getTime()
          });
        });

        it("should not be authenticated when session has not be initiated", function () {
          return test_non_authenticated_401(undefined);
        });

        it("should not be authenticated when domain is not allowed for user", function () {
          authSession.first_factor = true;
          authSession.second_factor = true;
          authSession.userid = "myuser";
          req.headers["x-original-url"] = "https://test.example.com/";
          mocks.accessController.isAccessAllowedMock.returns(false);

          return test_unauthorized_403({
            first_factor: true,
            second_factor: true,
            whitelisted: WhitelistValue.NOT_WHITELISTED,
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
        mocks.config.authentication_methods.per_subdomain_methods = {
          "redirect.url": "single_factor"
        };
      });

      it("should be authenticated when first factor is validated and second factor is not", function () {
        mocks.accessController.isAccessAllowedMock.returns(true);
        authSession.first_factor = true;
        authSession.userid = "user1";
        return VerifyGet.default(vars)(req as Express.Request, res as any)
          .then(function () {
            Assert(res.status.calledWith(204));
            Assert(res.send.calledOnce);
          });
      });

      it("should be rejected with 401 when first factor is not validated", function () {
        mocks.accessController.isAccessAllowedMock.returns(true);
        authSession.first_factor = false;
        return VerifyGet.default(vars)(req as Express.Request, res as any)
          .then(function () {
            Assert(res.status.calledWith(401));
          });
      });
    });

    describe("inactivity period", function () {
      it("should update last inactivity period on requests on /api/verify", function () {
        mocks.config.session.inactivity = 200000;
        mocks.accessController.isAccessAllowedMock.returns(true);
        const currentTime = new Date().getTime() - 1000;
        AuthenticationSessionHandler.reset(req as any);
        authSession.first_factor = true;
        authSession.second_factor = true;
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
        mocks.accessController.isAccessAllowedMock.returns(true);
        const currentTime = new Date().getTime() - 1000;
        AuthenticationSessionHandler.reset(req as any);
        authSession.first_factor = true;
        authSession.second_factor = true;
        authSession.userid = "myuser";
        authSession.groups = ["mygroup", "othergroup"];
        authSession.last_activity_datetime = currentTime;
        return VerifyGet.default(vars)(req as Express.Request, res as any)
          .then(function () {
            return AuthenticationSessionHandler.get(req as any, vars.logger);
          })
          .then(function (authSession) {
            Assert.equal(authSession.first_factor, false);
            Assert.equal(authSession.second_factor, false);
            Assert.equal(authSession.userid, undefined);
          });
      });
    });
  });

  describe("response type 401 | 302", function() {
    it("should return error code 401", function() {
      mocks.accessController.isAccessAllowedMock.returns(true);
      mocks.config.authentication_methods.default_method = "single_factor";
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
      mocks.accessController.isAccessAllowedMock.returns(true);
      mocks.config.authentication_methods.default_method = "single_factor";
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
      mocks.accessController.isAccessAllowedMock.returns(true);
      mocks.config.authentication_methods.default_method = "single_factor";
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
      mocks.accessController.isAccessAllowedMock.returns(true);
      mocks.config.authentication_methods.default_method = "single_factor";
      mocks.config.authentication_methods.per_subdomain_methods = {
        "secret.example.com": "two_factor"
      };
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
      mocks.accessController.isAccessAllowedMock.returns(true);
      mocks.config.authentication_methods.default_method = "single_factor";
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
      mocks.accessController.isAccessAllowedMock.returns(true);
      mocks.config.authentication_methods.default_method = "single_factor";
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
      mocks.accessController.isAccessAllowedMock.returns(true);
      mocks.config.authentication_methods.default_method = "single_factor";
      mocks.usersDatabase.checkUserPasswordStub.rejects(new Error(
        "Invalid credentials"));
      req.headers["proxy-authorization"] = "Basic am9objpwYXNzd29yZA==";

      return VerifyGet.default(vars)(req as Express.Request, res as any)
        .then(function () {
          Assert(res.status.calledWithExactly(401));
        });
    });

    it("should fail when resource is restricted", function () {
      mocks.accessController.isAccessAllowedMock.returns(false);
      mocks.config.authentication_methods.default_method = "single_factor";
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

