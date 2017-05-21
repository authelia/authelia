
import sinon = require("sinon");
import Promise = require("bluebird");
import assert = require("assert");
import u2f = require("../../../src/lib/routes/U2FRoutes");
import winston = require("winston");

import ExpressMock = require("../mocks/express");
import UserDataStoreMock = require("../mocks/UserDataStore");
import AuthdogMock = require("../mocks/authdog");

describe("test u2f routes", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let user_data_store: UserDataStoreMock.UserDataStore;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.app = {};
    req.app.get = sinon.stub();
    req.app.get.withArgs("logger").returns(winston);
    req.session = {};
    req.session.auth_session = {};
    req.session.auth_session.userid = "user";
    req.session.auth_session.first_factor = true;
    req.session.auth_session.second_factor = false;
    req.session.auth_session.identity_check = {};
    req.session.auth_session.identity_check.challenge = "u2f-register";
    req.session.auth_session.register_request = {};
    req.headers = {};
    req.headers.host = "localhost";

    const options = {
      inMemoryOnly: true
    };

    user_data_store = UserDataStoreMock.UserDataStore();
    user_data_store.set_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    user_data_store.get_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    req.app.get.withArgs("user data store").returns(user_data_store);

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  });

  describe("test registration request", test_registration_request);
  describe("test registration", test_registration);
  describe("test signing request", test_signing_request);
  describe("test signing", test_signing);

  function test_registration_request() {
    it("should send back the registration request and save it in the session", function (done) {
      const expectedRequest = {
        test: "abc"
      };
      res.json = sinon.spy(function (data: any) {
        assert.equal(200, res.status.getCall(0).args[0]);
        assert.deepEqual(expectedRequest, data);
        done();
      });
      const user_key_container = {};
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.startRegistration.returns(Promise.resolve(expectedRequest));

      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.register_request(req as any, res as any, undefined);
    });

    it("should return internal error on registration request", function (done) {
      res.send = sinon.spy(function (data: any) {
        assert.equal(500, res.status.getCall(0).args[0]);
        done();
      });
      const user_key_container = {};
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.startRegistration.returns(Promise.reject("Internal error"));

      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.register_request(req as any, res as any, undefined);
    });

    it("should return forbidden if identity has not been verified", function (done) {
      res.send = sinon.spy(function (data: any) {
        assert.equal(403, res.status.getCall(0).args[0]);
        done();
      });
      req.session.auth_session.identity_check = undefined;
      u2f.register_request(req as any, res as any, undefined);
    });
  }

  function test_registration() {
    it("should save u2f meta and return status code 200", function (done) {
      const expectedStatus = {
        keyHandle: "keyHandle",
        publicKey: "pbk",
        certificate: "cert"
      };
      res.send = sinon.spy(function (data: any) {
        assert.equal("user", user_data_store.set_u2f_meta.getCall(0).args[0]);
        assert.equal(req.session.auth_session.identity_check, undefined);
        done();
      });
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.finishRegistration.returns(Promise.resolve(expectedStatus));

      req.session.auth_session.register_request = {};
      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.register(req as any, res as any, undefined);
    });

    it("should return unauthorized on finishRegistration error", function (done) {
      res.send = sinon.spy(function (data: any) {
        assert.equal(500, res.status.getCall(0).args[0]);
        done();
      });
      const user_key_container = {};
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.finishRegistration.returns(Promise.reject("Internal error"));

      req.session.auth_session.register_request = "abc";
      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.register(req as any, res as any, undefined);
    });

    it("should return 403 when register_request is not provided", function (done) {
      res.send = sinon.spy(function (data: any) {
        assert.equal(403, res.status.getCall(0).args[0]);
        done();
      });
      const user_key_container = {};
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.finishRegistration.returns(Promise.resolve());

      req.session.auth_session.register_request = undefined;
      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.register(req as any, res as any, undefined);
    });

    it("should return forbidden error when no auth request has been initiated", function (done) {
      res.send = sinon.spy(function (data: any) {
        assert.equal(403, res.status.getCall(0).args[0]);
        done();
      });
      const user_key_container = {};
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.finishRegistration.returns(Promise.resolve());

      req.session.auth_session.register_request = undefined;
      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.register(req as any, res as any, undefined);
    });

    it("should return forbidden error when identity has not been verified", function (done) {
      res.send = sinon.spy(function (data: any) {
        assert.equal(403, res.status.getCall(0).args[0]);
        done();
      });
      req.session.auth_session.identity_check = undefined;
      u2f.register(req as any, res as any, undefined);
    });
  }

  function test_signing_request() {
    it("should send back the sign request and save it in the session", function (done) {
      const expectedRequest = {
        test: "abc"
      };
      res.json = sinon.spy(function (data: any) {
        assert.deepEqual(expectedRequest, req.session.auth_session.sign_request);
        assert.equal(200, res.status.getCall(0).args[0]);
        assert.deepEqual(expectedRequest, data);
        done();
      });
      const user_key_container = {
        user: {}
      };
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.startAuthentication.returns(Promise.resolve(expectedRequest));

      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.sign_request(req as any, res as any, undefined);
    });

    it("should return unauthorized error on registration request error", function (done) {
      res.send = sinon.spy(function (data: any) {
        assert.equal(500, res.status.getCall(0).args[0]);
        done();
      });
      const user_key_container = {
        user: {}
      };
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.startAuthentication.returns(Promise.reject("Internal error"));

      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.sign_request(req as any, res as any, undefined);
    });

    it("should send unauthorized error when no registration exists", function (done) {
      const expectedRequest = {
        test: "abc"
      };
      res.send = sinon.spy(function (data: any) {
        assert.equal(401, res.status.getCall(0).args[0]);
        done();
      });
      const user_key_container = {}; // no entry means no registration
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.startAuthentication.returns(Promise.resolve(expectedRequest));

      user_data_store.get_u2f_meta = sinon.stub().returns(Promise.resolve());

      req.app.get = sinon.stub();
      req.app.get.withArgs("logger").returns(winston);
      req.app.get.withArgs("user data store").returns(user_data_store);
      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.sign_request(req as any, res as any, undefined);
    });
  }

  function test_signing() {
    it("should return status code 204", function (done) {
      const user_key_container = {
        user: {}
      };
      const expectedStatus = {
        keyHandle: "keyHandle",
        publicKey: "pbk",
        certificate: "cert"
      };
      res.send = sinon.spy(function (data: any) {
        assert(204, res.status.getCall(0).args[0]);
        assert(req.session.auth_session.second_factor);
        done();
      });
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.finishAuthentication.returns(Promise.resolve(expectedStatus));

      req.session.auth_session.sign_request = {};
      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.sign(req as any, res as any, undefined);
    });

    it("should return unauthorized error on registration request internal error", function (done) {
      res.send = sinon.spy(function (data: any) {
        assert.equal(500, res.status.getCall(0).args[0]);
        done();
      });
      const user_key_container = {
        user: {}
      };

      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.finishAuthentication.returns(Promise.reject("Internal error"));

      req.session.auth_session.sign_request = {};
      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.sign(req as any, res as any, undefined);
    });

    it("should return unauthorized error when no sign request has been initiated", function (done) {
      res.send = sinon.spy(function (data: any) {
        assert.equal(401, res.status.getCall(0).args[0]);
        done();
      });
      const user_key_container = {};
      const u2f_mock = AuthdogMock.AuthdogMock();
      u2f_mock.finishAuthentication.returns(Promise.resolve());

      req.app.get.withArgs("u2f").returns(u2f_mock);
      u2f.sign(req as any, res as any, undefined);
    });
  }
});

