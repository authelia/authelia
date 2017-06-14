
import assert = require("assert");
import VerifyGet = require("../../../../src/server/lib/routes/verify/get");
import AuthenticationSession = require("../../../../src/server/lib/AuthenticationSession");

import sinon = require("sinon");
import winston = require("winston");
import BluebirdPromise = require("bluebird");

import express = require("express");

import ExpressMock = require("../../mocks/express");
import AccessControllerMock = require("../../mocks/AccessController");
import ServerVariablesMock = require("../../mocks/ServerVariablesMock");

describe("test authentication token verification", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let accessController: AccessControllerMock.AccessControllerMock;

  beforeEach(function () {
    accessController = AccessControllerMock.AccessControllerMock();
    accessController.isDomainAllowedForUser.returns(true);

    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
    req.headers = {};
    req.headers.host = "secret.example.com";
    const mocks = ServerVariablesMock.mock(req.app);
    mocks.config = {};
    mocks.logger = winston;
    mocks.accessController = accessController;
  });

  it("should be already authenticated", function (done) {
    req.session = {};
    AuthenticationSession.reset(req as any);
    const authSession = AuthenticationSession.get(req as any);
    authSession.first_factor = true;
    authSession.second_factor = true;
    authSession.userid = "myuser";

    res.send = sinon.spy(function () {
      assert.equal(204, res.status.getCall(0).args[0]);
      done();
    });

    VerifyGet.default(req as express.Request, res as any);
  });

  describe("given different cases of session", function () {
    function test_session(auth_session: AuthenticationSession.AuthenticationSession, status_code: number) {
      return new BluebirdPromise(function (resolve, reject) {
        req.session = {};
        req.session.auth_session = auth_session;

        res.send = sinon.spy(function () {
          assert.equal(status_code, res.status.getCall(0).args[0]);
          resolve();
        });

        VerifyGet.default(req as express.Request, res as any);
      });
    }

    function test_unauthorized(auth_session: AuthenticationSession.AuthenticationSession) {
      return test_session(auth_session, 401);
    }

    function test_authorized(auth_session: AuthenticationSession.AuthenticationSession) {
      return test_session(auth_session, 204);
    }

    it("should not be authenticated when second factor is missing", function () {
      return test_unauthorized({
        userid: "user",
        first_factor: true,
        second_factor: false,
        email: undefined,
        groups: [],
      });
    });

    it("should not be authenticated when first factor is missing", function () {
      return test_unauthorized({
        userid: "user",
        first_factor: false,
        second_factor: true,
        email: undefined,
        groups: [],
      });
    });

    it("should not be authenticated when userid is missing", function () {
      return test_unauthorized({
        userid: undefined,
        first_factor: true,
        second_factor: false,
        email: undefined,
        groups: [],
      });
    });

    it("should not be authenticated when first and second factor are missing", function () {
      return test_unauthorized({
        userid: "user",
        first_factor: false,
        second_factor: false,
        email: undefined,
        groups: [],
       });
    });

    it("should not be authenticated when session has not be initiated", function () {
      return test_unauthorized(undefined);
    });

    it("should not be authenticated when domain is not allowed for user", function () {
      req.headers.host = "test.example.com";

      accessController.isDomainAllowedForUser.returns(false);
      accessController.isDomainAllowedForUser.withArgs("test.example.com", "user", ["group1", "group2"]).returns(true);

      return test_unauthorized({
        first_factor: true,
        second_factor: true,
        userid: "user",
        groups: ["group1", "group2"],
        email: undefined
      });
    });
  });
});

