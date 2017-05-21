import sinon = require("sinon");
import winston = require("winston");
import u2f_register = require("../../../src/lib/routes/U2FRegistration");
import assert = require("assert");

import ExpressMock = require("../mocks/express");
import UserDataStoreMock = require("../mocks/UserDataStore");

describe("test register handler", function() {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let user_data_store: UserDataStoreMock.UserDataStore;

  beforeEach(function() {
    req = ExpressMock.RequestMock;
    req.app = {};
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

    user_data_store = UserDataStoreMock.UserDataStore();
    user_data_store.set_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    user_data_store.get_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    user_data_store.issue_identity_check_token = sinon.stub().returns(Promise.resolve({}));
    user_data_store.consume_identity_check_token = sinon.stub().returns(Promise.resolve({}));
    req.app.get.withArgs("user data store").returns(user_data_store);

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  });

  describe("test u2f registration check", test_registration_check);

  function test_registration_check() {
    it("should fail if first_factor has not been passed", function(done) {
      req.session.auth_session.first_factor = false;
      u2f_register.icheck_interface.preValidation(req as any)
      .catch(function(err: Error) {
        done();
      });
    });

    it("should fail if userid is missing", function(done) {
      req.session.auth_session.first_factor = false;
      req.session.auth_session.userid = undefined;

      u2f_register.icheck_interface.preValidation(req as any)
      .catch(function(err: Error) {
        done();
      });
    });

    it("should fail if email is missing", function(done) {
      req.session.auth_session.first_factor = false;
      req.session.auth_session.email = undefined;

      u2f_register.icheck_interface.preValidation(req as any)
      .catch(function(err) {
        done();
      });
    });

    it("should succeed if first factor passed, userid and email are provided", function(done) {
      u2f_register.icheck_interface.preValidation(req as any)
      .then(function(err) {
        done();
      });
    });
  }
});
