
import Assert = require("assert");
import VerifyGet = require("../../../src/lib/routes/verify/get");
import AuthenticationSession = require("../../../src/lib/AuthenticationSession");
import { AuthenticationMethodCalculator } from "../../../src/lib/AuthenticationMethodCalculator";
import { AuthenticationMethodsConfiguration } from "../../../src/lib/configuration/Configuration";
import Sinon = require("sinon");
import winston = require("winston");
import BluebirdPromise = require("bluebird");
import express = require("express");
import ExpressMock = require("../../mocks/express");
import { ServerVariables } from "../../../src/lib/ServerVariables";
import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../mocks/ServerVariablesMockBuilder";

describe("test /verify endpoint", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
    req.session = {};
    req.query = {
      redirect: "http://redirect.url"
    };
    req.app = {
      get: Sinon.stub().returns({ logger: winston })
    };
    AuthenticationSession.reset(req as any);
    req.headers = {};
    req.headers.host = "secret.example.com";
    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;
  });

  it("should be already authenticated", function () {
    req.session = {};
    mocks.accessController.isAccessAllowedMock.returns(true);
    AuthenticationSession.reset(req as any);
    return AuthenticationSession.get(req as any)
      .then(function (authSession: AuthenticationSession.AuthenticationSession) {
        authSession.first_factor = true;
        authSession.second_factor = true;
        authSession.userid = "myuser";
        authSession.groups = ["mygroup", "othergroup"];
        return VerifyGet.default(vars)(req as express.Request, res as any);
      })
      .then(function () {
        Sinon.assert.calledWithExactly(res.setHeader, "Remote-User", "myuser");
        Sinon.assert.calledWithExactly(res.setHeader, "Remote-Groups", "mygroup,othergroup");
        Assert.equal(204, res.status.getCall(0).args[0]);
      });
  });

  function test_session(_authSession: AuthenticationSession.AuthenticationSession, status_code: number) {
    return AuthenticationSession.get(req as any)
      .then(function (authSession) {
        authSession = _authSession;
        return VerifyGet.default(vars)(req as express.Request, res as any);
      })
      .then(function () {
        Assert.equal(status_code, res.status.getCall(0).args[0]);
      });
  }

  function test_non_authenticated_401(authSession: AuthenticationSession.AuthenticationSession) {
    return test_session(authSession, 401);
  }

  function test_unauthorized_403(authSession: AuthenticationSession.AuthenticationSession) {
    return test_session(authSession, 403);
  }

  function test_authorized(authSession: AuthenticationSession.AuthenticationSession) {
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
          email: undefined,
          groups: [],
          last_activity_datetime: new Date().getTime()
        });
      });

      it("should not be authenticated when session has not be initiated", function () {
        return test_non_authenticated_401(undefined);
      });

      it("should not be authenticated when domain is not allowed for user", function () {
        return AuthenticationSession.get(req as any)
          .then(function (authSession) {
            authSession.first_factor = true;
            authSession.second_factor = true;
            authSession.userid = "myuser";
            req.headers.host = "test.example.com";
            mocks.accessController.isAccessAllowedMock.returns(false);

            return test_unauthorized_403({
              first_factor: true,
              second_factor: true,
              userid: "user",
              groups: ["group1", "group2"],
              email: undefined,
              last_activity_datetime: new Date().getTime()
            });
          });
      });
    });
  });

  describe("given user tries to access a basic auth endpoint", function () {
    beforeEach(function () {
      req.query = {
        redirect: "http://redirect.url"
      };
      req.headers["host"] = "redirect.url";
      mocks.config.authentication_methods.per_subdomain_methods = {
        "redirect.url": "basic_auth"
      };
    });

    it("should be authenticated when first factor is validated and second factor is not", function () {
      mocks.accessController.isAccessAllowedMock.returns(true);
      return AuthenticationSession.get(req as any)
        .then(function (authSession) {
          authSession.first_factor = true;
          authSession.userid = "user1";
          return VerifyGet.default(vars)(req as express.Request, res as any);
        })
        .then(function () {
          Assert(res.status.calledWith(204));
          Assert(res.send.calledOnce);
        });
    });

    it("should be rejected with 401 when first factor is not validated", function () {
      mocks.accessController.isAccessAllowedMock.returns(true);
      return AuthenticationSession.get(req as any)
        .then(function (authSession) {
          authSession.first_factor = false;
          return VerifyGet.default(vars)(req as express.Request, res as any);
        })
        .then(function () {
          Assert(res.status.calledWith(401));
        });
    });
  });

  describe("inactivity period", function () {
    it("should update last inactivity period on requests on /verify", function () {
      mocks.config.session.inactivity = 200000;
      mocks.accessController.isAccessAllowedMock.returns(true);
      const currentTime = new Date().getTime() - 1000;
      AuthenticationSession.reset(req as any);
      return AuthenticationSession.get(req as any)
        .then(function (authSession: AuthenticationSession.AuthenticationSession) {
          authSession.first_factor = true;
          authSession.second_factor = true;
          authSession.userid = "myuser";
          authSession.groups = ["mygroup", "othergroup"];
          authSession.last_activity_datetime = currentTime;
          return VerifyGet.default(vars)(req as express.Request, res as any);
        })
        .then(function () {
          return AuthenticationSession.get(req as any);
        })
        .then(function (authSession) {
          Assert(authSession.last_activity_datetime > currentTime);
        });
    });

    it("should reset session when max inactivity period has been reached", function () {
      mocks.config.session.inactivity = 1;
      mocks.accessController.isAccessAllowedMock.returns(true);
      const currentTime = new Date().getTime() - 1000;
      AuthenticationSession.reset(req as any);
      return AuthenticationSession.get(req as any)
        .then(function (authSession: AuthenticationSession.AuthenticationSession) {
          authSession.first_factor = true;
          authSession.second_factor = true;
          authSession.userid = "myuser";
          authSession.groups = ["mygroup", "othergroup"];
          authSession.last_activity_datetime = currentTime;
          return VerifyGet.default(vars)(req as express.Request, res as any);
        })
        .then(function () {
          return AuthenticationSession.get(req as any);
        })
        .then(function (authSession) {
          Assert.equal(authSession.first_factor, false);
          Assert.equal(authSession.second_factor, false);
          Assert.equal(authSession.userid, undefined);
        });
    });
  });
});

