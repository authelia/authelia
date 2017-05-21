import sinon = require("sinon");
import winston = require("winston");
import TOTPRegistration = require("../../../src/lib/routes/TOTPRegistration");
import assert = require("assert");
import BluebirdPromise = require("bluebird");

import ExpressMock = require("../mocks/express");
import UserDataStoreMock = require("../mocks/UserDataStore");

describe("test totp register", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let userDataStore: UserDataStoreMock.UserDataStore;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.app.get = sinon.stub();
    req.app.get.withArgs("logger").returns(winston);
    req.session = {};
    req.session.auth_session = {};
    req.session.auth_session.userid = "user";
    req.session.auth_session.email = "user@example.com";
    req.session.auth_session.first_factor = true;
    req.session.auth_session.second_factor = false;
    req.headers = {};
    req.headers.host = "localhost";

    const options = {
      inMemoryOnly: true
    };

    userDataStore = UserDataStoreMock.UserDataStore();
    userDataStore.set_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    userDataStore.get_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    userDataStore.issue_identity_check_token = sinon.stub().returns(Promise.resolve({}));
    userDataStore.consume_identity_check_token = sinon.stub().returns(Promise.resolve({}));
    userDataStore.set_totp_secret = sinon.stub().returns(Promise.resolve({}));
    req.app.get.withArgs("user data store").returns(userDataStore);

    res = ExpressMock.ResponseMock();
  });

  describe("test totp registration check", test_registration_check);
  describe("test totp post secret", test_post_secret);

  function test_registration_check() {
    it("should fail if first_factor has not been passed", function (done) {
      req.session.auth_session.first_factor = false;
      TOTPRegistration.icheck_interface.preValidation(req as any)
        .catch(function (err) {
          done();
        });
    });

    it("should fail if userid is missing", function (done) {
      req.session.auth_session.first_factor = false;
      req.session.auth_session.userid = undefined;

      TOTPRegistration.icheck_interface.preValidation(req as any)
        .catch(function (err) {
          done();
        });
    });

    it("should fail if email is missing", function (done) {
      req.session.auth_session.first_factor = false;
      req.session.auth_session.email = undefined;

      TOTPRegistration.icheck_interface.preValidation(req as any)
        .catch(function (err) {
          done();
        });
    });

    it("should succeed if first factor passed, userid and email are provided", function (done) {
      TOTPRegistration.icheck_interface.preValidation(req as any)
        .then(function (err) {
          done();
        });
    });
  }

  function test_post_secret() {
    it("should send the secret in json format", function (done) {
      req.app.get.withArgs("totp generator").returns({
        generate: sinon.stub().returns({ otpauth_url: "abc" })
      });
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.userid = "user";
      req.session.auth_session.identity_check.challenge = "totp-register";
      res.json = sinon.spy(function () {
        done();
      });
      TOTPRegistration.post(req as any, res as any);
    });

    it("should clear the session for reauthentication", function (done) {
      req.app.get.withArgs("totp generator").returns({
        generate: sinon.stub().returns({ otpauth_url: "abc" })
      });
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.userid = "user";
      req.session.auth_session.identity_check.challenge = "totp-register";
      res.json = sinon.spy(function () {
        assert.equal(req.session, undefined);
        done();
      });
      TOTPRegistration.post(req as any, res as any);
    });

    it("should return 403 if the identity check challenge is not set", function (done) {
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.challenge = undefined;
      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });
      TOTPRegistration.post(req as any, res as any);
    });

    it("should return 500 if db throws", function (done) {
      req.app.get.withArgs("totp generator").returns({
        generate: sinon.stub().returns({ otpauth_url: "abc" })
      });
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.userid = "user";
      req.session.auth_session.identity_check.challenge = "totp-register";
      userDataStore.set_totp_secret.returns(BluebirdPromise.reject("internal error"));

      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 500);
        done();
      });
      TOTPRegistration.post(req as any, res as any);
    });
  }
});
