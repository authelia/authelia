
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import U2FRegisterPost = require("../../../../../../src/server/lib/routes/secondfactor/u2f/register/post");
import AuthenticationSession = require("../../../../../../src/server/lib/AuthenticationSession");
import winston = require("winston");

import ExpressMock = require("../../../../mocks/express");
import UserDataStoreMock = require("../../../../mocks/UserDataStore");
import U2FMock = require("../../../../mocks/u2f");
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");
import U2f = require("u2f");

describe("test u2f routes: register", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let userDataStore: UserDataStoreMock.UserDataStore;
  let mocks: ServerVariablesMock.ServerVariablesMock;
  let authSession: AuthenticationSession.AuthenticationSession;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.app = {};
    mocks = ServerVariablesMock.mock(req.app);
    mocks.logger = winston;

    req.session = {};
    AuthenticationSession.reset(req as any);
    authSession = AuthenticationSession.get(req as any);
    authSession.userid = "user";
    authSession.first_factor = true;
    authSession.second_factor = false;
    authSession.identity_check = {
      challenge: "u2f-register",
      userid: "user"
    };

    req.headers = {};
    req.headers.host = "localhost";

    const options = {
      inMemoryOnly: true
    };

    userDataStore = UserDataStoreMock.UserDataStore();
    userDataStore.set_u2f_meta = sinon.stub().returns(BluebirdPromise.resolve({}));
    userDataStore.get_u2f_meta = sinon.stub().returns(BluebirdPromise.resolve({}));
    mocks.userDataStore = userDataStore;

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
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

      authSession.register_request = {
        appId: "app",
        challenge: "challenge",
        keyHandle: "key",
        version: "U2F_V2"
      };
      mocks.u2f = u2f_mock;
      return U2FRegisterPost.default(req as any, res as any)
        .then(function () {
          assert.equal("user", userDataStore.set_u2f_meta.getCall(0).args[0]);
          assert.equal(authSession.identity_check, undefined);
        });
    });

    it("should return unauthorized on finishRegistration error", function () {
      const user_key_container = {};
      const u2f_mock = U2FMock.U2FMock();
      u2f_mock.checkRegistration.returns({ errorCode: 500 });

      authSession.register_request = {
        appId: "app",
        challenge: "challenge",
        keyHandle: "key",
        version: "U2F_V2"
      };
      mocks.u2f = u2f_mock;
      return U2FRegisterPost.default(req as any, res as any)
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

      authSession.register_request = undefined;
      mocks.u2f = u2f_mock;
      return U2FRegisterPost.default(req as any, res as any)
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

      authSession.register_request = undefined;
      mocks.u2f = u2f_mock;
      return U2FRegisterPost.default(req as any, res as any)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function () {
          assert.equal(403, res.status.getCall(0).args[0]);
          return BluebirdPromise.resolve();
        });
    });

    it("should return forbidden error when identity has not been verified", function () {
      authSession.identity_check = undefined;
      return U2FRegisterPost.default(req as any, res as any)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function () {
          assert.equal(403, res.status.getCall(0).args[0]);
          return BluebirdPromise.resolve();
        });
    });
  }
});

