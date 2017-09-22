
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import U2FRegisterPost = require("../../../../../../../src/server/lib/routes/secondfactor/u2f/register/post");
import AuthenticationSession = require("../../../../../../../src/server/lib/AuthenticationSession");
import { ServerVariablesHandler } from "../../../../../../../src/server/lib/ServerVariablesHandler";
import winston = require("winston");

import ExpressMock = require("../../../../mocks/express");
import { UserDataStoreStub } from "../../../../mocks/storage/UserDataStoreStub";
import U2FMock = require("../../../../mocks/u2f");
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");
import U2f = require("u2f");

describe("test u2f routes: register", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock.ServerVariablesMock;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.app = {};
    mocks = ServerVariablesMock.mock(req.app);
    mocks.logger = winston;

    req.session = {};
    req.headers = {};
    req.headers.host = "localhost";

    const options = {
      inMemoryOnly: true
    };

    mocks.userDataStore.saveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();

    AuthenticationSession.reset(req as any);
    return AuthenticationSession.get(req as any)
      .then(function (authSession: AuthenticationSession.AuthenticationSession) {
        authSession.userid = "user";
        authSession.first_factor = true;
        authSession.second_factor = false;
        authSession.identity_check = {
          challenge: "u2f-register",
          userid: "user"
        };
      });
  });

  describe("test registration", test_registration);


  function test_registration() {
    it("should save u2f meta and return status code 200", function () {
      const expectedStatus = {
        keyHandle: "keyHandle",
        publicKey: "pbk",
        certificate: "cert"
      };
      const u2f_mock = U2FMock.U2FMock();
      u2f_mock.checkRegistration.returns(BluebirdPromise.resolve(expectedStatus));
      mocks.u2f = u2f_mock;

      return AuthenticationSession.get(req as any)
        .then(function (authSession) {
          authSession.register_request = {
            appId: "app",
            challenge: "challenge",
            keyHandle: "key",
            version: "U2F_V2"
          };
          return U2FRegisterPost.default(req as any, res as any);
        })
        .then(function () {
          return AuthenticationSession.get(req as any);
        })
        .then(function (authSession) {
          assert.equal("user", mocks.userDataStore.saveU2FRegistrationStub.getCall(0).args[0]);
          assert.equal(authSession.identity_check, undefined);
        });
    });

    it("should return unauthorized on finishRegistration error", function () {
      const user_key_container = {};
      const u2f_mock = U2FMock.U2FMock();
      u2f_mock.checkRegistration.returns({ errorCode: 500 });
      mocks.u2f = u2f_mock;

      return AuthenticationSession.get(req as any)
        .then(function (authSession) {
          authSession.register_request = {
            appId: "app",
            challenge: "challenge",
            keyHandle: "key",
            version: "U2F_V2"
          };

          return U2FRegisterPost.default(req as any, res as any);
        })
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function () {
          assert.equal(500, res.status.getCall(0).args[0]);
          return BluebirdPromise.resolve();
        });
    });

    it("should return 403 when register_request is not provided", function () {
      const user_key_container = {};
      const u2f_mock = U2FMock.U2FMock();
      u2f_mock.checkRegistration.returns(BluebirdPromise.resolve());

      mocks.u2f = u2f_mock;
      return AuthenticationSession.get(req as any)
        .then(function (authSession) {
          authSession.register_request = undefined;
          return U2FRegisterPost.default(req as any, res as any);
        })
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function () {
          assert.equal(403, res.status.getCall(0).args[0]);
          return BluebirdPromise.resolve();
        });
    });

    it("should return forbidden error when no auth request has been initiated", function () {
      const user_key_container = {};
      const u2f_mock = U2FMock.U2FMock();
      u2f_mock.checkRegistration.returns(BluebirdPromise.resolve());

      mocks.u2f = u2f_mock;
      return AuthenticationSession.get(req as any)
        .then(function (authSession) {
          authSession.register_request = undefined;
          return U2FRegisterPost.default(req as any, res as any);
        })
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function () {
          assert.equal(403, res.status.getCall(0).args[0]);
          return BluebirdPromise.resolve();
        });
    });

    it("should return forbidden error when identity has not been verified", function () {
      return AuthenticationSession.get(req as any)
        .then(function (authSession) {
          authSession.identity_check = undefined;
          return U2FRegisterPost.default(req as any, res as any);
        })
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function () {
          assert.equal(403, res.status.getCall(0).args[0]);
          return BluebirdPromise.resolve();
        });
    });
  }
});

