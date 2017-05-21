
import BluebirdPromise = require("bluebird");
import sinon = require("sinon");
import assert = require("assert");
import winston = require("winston");

import exceptions = require("../../../src/lib/Exceptions");
import TOTPAuthenticator = require("../../../src/lib/routes/TOTPAuthenticator");

import ExpressMock = require("../mocks/express");
import UserDataStoreMock = require("../mocks/UserDataStore");
import TOTPValidatorMock = require("../mocks/TOTPValidator");

describe("test totp route", function() {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let totpValidator: TOTPValidatorMock.TOTPValidatorMock;
  let userDataStore: UserDataStoreMock.UserDataStore;

  beforeEach(function() {
    const app_get = sinon.stub();
    req = {
      app: {
        get: app_get
      },
      body: {
        token: "abc"
      },
      session: {
        auth_session: {
          userid: "user",
          first_factor: false,
          second_factor: false
        }
      }
    };
    res = ExpressMock.ResponseMock();

    const config = { totp_secret: "secret" };
    totpValidator = TOTPValidatorMock.TOTPValidatorMock();

    userDataStore = UserDataStoreMock.UserDataStore();

    const doc = {
      userid: "user",
      secret: {
        base32: "ABCDEF"
      }
    };
    userDataStore.get_totp_secret.returns(BluebirdPromise.resolve(doc));

    app_get.withArgs("logger").returns(winston);
    app_get.withArgs("totp validator").returns(totpValidator);
    app_get.withArgs("config").returns(config);
    app_get.withArgs("user data store").returns(userDataStore);
  });


  it("should send status code 204 when totp is valid", function(done) {
    totpValidator.validate.returns(Promise.resolve("ok"));
    res.send = sinon.spy(function() {
      // Second factor passed
      assert.equal(true, req.session.auth_session.second_factor);
      assert.equal(204, res.status.getCall(0).args[0]);
      done();
    });
    TOTPAuthenticator(req as any, res as any);
  });

  it("should send status code 401 when totp is not valid", function(done) {
    totpValidator.validate.returns(Promise.reject(new exceptions.InvalidTOTPError("Bad TOTP token")));
    res.send = sinon.spy(function() {
      assert.equal(false, req.session.auth_session.second_factor);
      assert.equal(401, res.status.getCall(0).args[0]);
      done();
    });
    TOTPAuthenticator(req as any, res as any);
  });

  it("should send status code 401 when session has not been initiated", function(done) {
    totpValidator.validate.returns(Promise.resolve("abc"));
    res.send = sinon.spy(function() {
      assert.equal(403, res.status.getCall(0).args[0]);
      done();
    });
    req.session = {};
    TOTPAuthenticator(req as any, res as any);
  });
});

