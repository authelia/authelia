
import Assert = require("assert");
import VerifyGet = require("../../../src/lib/routes/verify/get");
import AuthenticationSession = require("../../../src/lib/AuthenticationSession");

import Sinon = require("sinon");
import winston = require("winston");
import BluebirdPromise = require("bluebird");

import express = require("express");

import ExpressMock = require("../../mocks/express");
import { AccessControllerStub } from "../../mocks/AccessControllerStub";
import ServerVariablesMock = require("../../mocks/ServerVariablesMock");

describe("test authentication token verification", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let accessController: AccessControllerStub;

  beforeEach(function () {
    accessController = new AccessControllerStub();
    accessController.isAccessAllowedMock.returns(true);

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
    const mocks = ServerVariablesMock.mock(req.app);
    mocks.config = {} as any;
    mocks.logger = winston;
    mocks.accessController = accessController as any;
  });

  it("should be already authenticated", function () {
    req.session = {};
    AuthenticationSession.reset(req as any);
    return AuthenticationSession.get(req as any)
      .then(function (authSession: AuthenticationSession.AuthenticationSession) {
        authSession.first_factor = true;
        authSession.second_factor = true;
        authSession.userid = "myuser";
        authSession.groups = ["mygroup", "othergroup"];
        return VerifyGet.default(req as express.Request, res as any);
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
        return VerifyGet.default(req as express.Request, res as any);
      })
      .then(function () {
        Assert.equal(status_code, res.status.getCall(0).args[0]);
      });
  }

  function test_non_authenticated_401(auth_session: AuthenticationSession.AuthenticationSession) {
    return test_session(auth_session, 401);
  }

  function test_unauthorized_403(auth_session: AuthenticationSession.AuthenticationSession) {
    return test_session(auth_session, 403);
  }

  function test_authorized(auth_session: AuthenticationSession.AuthenticationSession) {
    return test_session(auth_session, 204);
  }

  describe("given user tries to access a 2-factor endpoint", function () {
    describe("given different cases of session", function () {
      it("should not be authenticated when second factor is missing", function () {
        return test_non_authenticated_401({
          userid: "user",
          first_factor: true,
          second_factor: false,
          email: undefined,
          groups: [],
        });
      });

      it("should not be authenticated when first factor is missing", function () {
        return test_non_authenticated_401({
          userid: "user",
          first_factor: false,
          second_factor: true,
          email: undefined,
          groups: [],
        });
      });

      it("should not be authenticated when userid is missing", function () {
        return test_non_authenticated_401({
          userid: undefined,
          first_factor: true,
          second_factor: false,
          email: undefined,
          groups: [],
        });
      });

      it("should not be authenticated when first and second factor are missing", function () {
        return test_non_authenticated_401({
          userid: "user",
          first_factor: false,
          second_factor: false,
          email: undefined,
          groups: [],
        });
      });

      it("should not be authenticated when session has not be initiated", function () {
        return test_non_authenticated_401(undefined);
      });

      it("should not be authenticated when domain is not allowed for user", function () {
        return AuthenticationSession.get(req as any)
          .then(function (authSession: AuthenticationSession.AuthenticationSession) {
            authSession.first_factor = true;
            authSession.second_factor = true;
            authSession.userid = "myuser";

            req.headers.host = "test.example.com";

            accessController.isAccessAllowedMock.returns(false);
            accessController.isAccessAllowedMock.withArgs("test.example.com", "user", ["group1", "group2"]).returns(true);

            return test_unauthorized_403({
              first_factor: true,
              second_factor: true,
              userid: "user",
              groups: ["group1", "group2"],
              email: undefined
            });
          });
      });
    });
  });

  describe("given user tries to access a basic auth endpoint", function () {
    beforeEach(function () {
      req.query = {
        redirect: "http://redirect.url",
        only_basic_auth: "true"
      };
    });

    it("should be authenticated when first factor is validated and not second factor", function () {
      return AuthenticationSession.get(req as any)
        .then(function (authSession: AuthenticationSession.AuthenticationSession) {
          authSession.first_factor = true;
          return VerifyGet.default(req as express.Request, res as any);
        })
        .then(function () {
          Assert(res.status.calledWith(204));
          Assert(res.send.calledOnce);
        });
    });

    it("should be rejected with 401 when first factor is not validated", function () {
      return AuthenticationSession.get(req as any)
        .then(function (authSession: AuthenticationSession.AuthenticationSession) {
          authSession.first_factor = false;
          return VerifyGet.default(req as express.Request, res as any);
        })
        .then(function () {
          Assert(res.status.calledWith(401));
        });
    });
  });
});

