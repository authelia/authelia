
import BluebirdPromise = require("bluebird");
import sinon = require("sinon");
import assert = require("assert");
import winston = require("winston");

import exceptions = require("../../../../../../src/server/lib/Exceptions");
import AuthenticationSession = require("../../../../../../src/server/lib/AuthenticationSession");
import SignPost = require("../../../../../../src/server/lib/routes/secondfactor/totp/sign/post");

import ExpressMock = require("../../../../mocks/express");
import UserDataStoreMock = require("../../../../mocks/UserDataStore");
import TOTPValidatorMock = require("../../../../mocks/TOTPValidator");
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");

describe("test totp route", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let totpValidator: TOTPValidatorMock.TOTPValidatorMock;
  let userDataStore: UserDataStoreMock.UserDataStore;
  let authSession: AuthenticationSession.AuthenticationSession;

  beforeEach(function () {
    const app_get = sinon.stub();
    req = {
      app: {
      },
      body: {
        token: "abc"
      },
      session: {}
    };
    AuthenticationSession.reset(req as any);
    authSession = AuthenticationSession.get(req as any);
    authSession.userid = "user";
    authSession.first_factor = true;
    authSession.second_factor = false;

    const mocks = ServerVariablesMock.mock(req.app);
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

    mocks.logger = winston;
    mocks.totpValidator = totpValidator;
    mocks.config = config;
    mocks.userDataStore = userDataStore;
  });


  it("should send status code 200 when totp is valid", function () {
    totpValidator.validate.returns(BluebirdPromise.resolve("ok"));
    return SignPost.default(req as any, res as any)
      .then(function () {
        assert.equal(true, authSession.second_factor);
        return BluebirdPromise.resolve();
      });
  });

  it("should send status code 401 when totp is not valid", function () {
    totpValidator.validate.returns(BluebirdPromise.reject(new exceptions.InvalidTOTPError("Bad TOTP token")));
    SignPost.default(req as any, res as any)
      .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
      .catch(function () {
        assert.equal(false, authSession.second_factor);
        assert.equal(401, res.status.getCall(0).args[0]);
        return BluebirdPromise.resolve();
      });
  });

  it("should send status code 401 when session has not been initiated", function () {
    totpValidator.validate.returns(BluebirdPromise.resolve("abc"));
    req.session = {};
    return SignPost.default(req as any, res as any)
      .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
      .catch(function () {
        assert.equal(401, res.status.getCall(0).args[0]);
        return BluebirdPromise.resolve();
      });
  });
});

