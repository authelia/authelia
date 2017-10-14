
import BluebirdPromise = require("bluebird");
import sinon = require("sinon");
import assert = require("assert");
import winston = require("winston");

import exceptions = require("../../../../../src/lib/Exceptions");
import AuthenticationSession = require("../../../../../src/lib/AuthenticationSession");
import SignPost = require("../../../../../src/lib/routes/secondfactor/totp/sign/post");

import ExpressMock = require("../../../../mocks/express");
import TOTPValidatorMock = require("../../../../mocks/TOTPValidator");
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");
import { UserDataStoreStub } from "../../../../mocks/storage/UserDataStoreStub";

describe("test totp route", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let totpValidator: TOTPValidatorMock.TOTPValidatorMock;
  let authSession: AuthenticationSession.AuthenticationSession;

  beforeEach(function () {
    const app_get = sinon.stub();
    req = {
      app: {
        get: sinon.stub().returns({ logger: winston })
      },
      body: {
        token: "abc"
      },
      session: {}
    };
    AuthenticationSession.reset(req as any);
    const mocks = ServerVariablesMock.mock(req.app);
    res = ExpressMock.ResponseMock();

    const config = { totp_secret: "secret" };
    totpValidator = TOTPValidatorMock.TOTPValidatorMock();

    const doc = {
      userid: "user",
      secret: {
        base32: "ABCDEF"
      }
    };
    mocks.userDataStore.retrieveTOTPSecretStub.returns(BluebirdPromise.resolve(doc));
    mocks.totpValidator = totpValidator;
    mocks.config = config;

    return AuthenticationSession.get(req as any)
      .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
        authSession = _authSession;
        authSession.userid = "user";
        authSession.first_factor = true;
        authSession.second_factor = false;
      });
  });


  it("should send status code 200 when totp is valid", function () {
    totpValidator.validate.returns(BluebirdPromise.resolve("ok"));
    return SignPost.default(req as any, res as any)
      .then(function () {
        assert.equal(true, authSession.second_factor);
        return BluebirdPromise.resolve();
      });
  });

  it("should send error message when totp is not valid", function () {
    totpValidator.validate.returns(BluebirdPromise.reject(new exceptions.InvalidTOTPError("Bad TOTP token")));
    return SignPost.default(req as any, res as any)
      .then(function () {
        assert.equal(false, authSession.second_factor);
        assert.equal(res.status.getCall(0).args[0], 200);
        assert.deepEqual(res.send.getCall(0).args[0], {
          error: "Operation failed."
        });
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

