
import assert = require("assert");
import AuthenticationValidator = require("../../../src/lib/routes/AuthenticationValidator");
import sinon = require("sinon");
import winston = require("winston");

import express = require("express");

import ExpressMock = require("../mocks/express");
import AccessControllerMock = require("../mocks/AccessController");

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
    req.app.get = sinon.stub();
    req.app.get.withArgs("config").returns({});
    req.app.get.withArgs("logger").returns(winston);
    req.app.get.withArgs("access controller").returns(accessController);
  });

  interface AuthenticationSession {
    first_factor?: boolean;
    second_factor?: boolean;
    userid?: string;
    groups?: string[];
  }

  it("should be already authenticated", function (done) {
    req.session = {};
    req.session.auth_session = {
      first_factor: true,
      second_factor: true,
      userid: "myuser",
    } as AuthenticationSession;

    res.send = sinon.spy(function () {
      assert.equal(204, res.status.getCall(0).args[0]);
      done();
    });

    AuthenticationValidator(req as express.Request, res as any);
  });

  describe("given different cases of session", function () {
    function test_session(auth_session: AuthenticationSession, status_code: number) {
      return new Promise(function (resolve, reject) {
        req.session = {};
        req.session.auth_session = auth_session;

        res.send = sinon.spy(function () {
          assert.equal(status_code, res.status.getCall(0).args[0]);
          resolve();
        });

        AuthenticationValidator(req as express.Request, res as any);
      });
    }

    function test_unauthorized(auth_session: AuthenticationSession) {
      return test_session(auth_session, 401);
    }

    function test_authorized(auth_session: AuthenticationSession) {
      return test_session(auth_session, 204);
    }

    it("should not be authenticated when second factor is missing", function () {
      return test_unauthorized({
        userid: "user",
        first_factor: true,
        second_factor: false
      });
    });

    it("should not be authenticated when first factor is missing", function () {
      return test_unauthorized({ first_factor: false, second_factor: true });
    });

    it("should not be authenticated when userid is missing", function () {
      return test_unauthorized({
        first_factor: true,
        second_factor: true,
        groups: ["mygroup"],
      });
    });

    it("should not be authenticated when first and second factor are missing", function () {
      return test_unauthorized({ first_factor: false, second_factor: false });
    });

    it("should not be authenticated when session has not be initiated", function () {
      return test_unauthorized(undefined);
    });

    it("should not be authenticated when session is partially initialized", function () {
      return test_unauthorized({ first_factor: true });
    });

    it.only("should not be authenticated when domain is not allowed for user", function () {
      req.headers.host = "test.example.com";

      accessController.isDomainAllowedForUser.returns(false);
      accessController.isDomainAllowedForUser.withArgs("test.example.com", "user", ["group1", "group2"]).returns(true);

      return test_authorized({
        first_factor: true,
        second_factor: true,
        userid: "user",
        groups: ["group1", "group2"]
      });
    });
  });
});

