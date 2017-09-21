import Sinon = require("sinon");
import winston = require("winston");
import RegistrationHandler from "../../../../../../../src/server/lib/routes/secondfactor/totp/identity/RegistrationHandler";
import { Identity } from "../../../../../../../src/types/Identity";
import AuthenticationSession = require("../../../../../../../src/server/lib/AuthenticationSession");
import { UserDataStore } from "../../../../../../../src/server/lib/storage/UserDataStore";
import assert = require("assert");
import BluebirdPromise = require("bluebird");

import ExpressMock = require("../../../../mocks/express");
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");

describe("test totp register", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  const registrationHandler: RegistrationHandler = new RegistrationHandler();
  let authSession: AuthenticationSession.AuthenticationSession;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    const mocks = ServerVariablesMock.mock(req.app);
    mocks.logger = winston;
    req.session = {};

    AuthenticationSession.reset(req as any);

    req.headers = {};
    req.headers.host = "localhost";

    const options = {
      inMemoryOnly: true
    };

    mocks.userDataStore.saveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.produceIdentityValidationTokenStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.consumeIdentityValidationTokenStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.saveTOTPSecretStub.returns(BluebirdPromise.resolve({}));

    res = ExpressMock.ResponseMock();

    return AuthenticationSession.get(req as any)
      .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
        authSession = _authSession;
        authSession.userid = "user";
        authSession.email = "user@example.com";
        authSession.first_factor = true;
        authSession.second_factor = false;
      });
  });

  describe("test totp registration check", test_registration_check);

  function test_registration_check() {
    it("should fail if first_factor has not been passed", function () {
      authSession.first_factor = false;
      return registrationHandler.preValidationInit(req as any)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function (err: Error) {
          return BluebirdPromise.resolve();
        });
    });

    it("should fail if userid is missing", function (done) {
      authSession.first_factor = false;
      authSession.userid = undefined;

      registrationHandler.preValidationInit(req as any)
        .catch(function (err: Error) {
          done();
        });
    });

    it("should fail if email is missing", function (done) {
      authSession.first_factor = false;
      authSession.email = undefined;

      registrationHandler.preValidationInit(req as any)
        .catch(function (err: Error) {
          done();
        });
    });

    it("should succeed if first factor passed, userid and email are provided", function (done) {
      registrationHandler.preValidationInit(req as any)
        .then(function (identity: Identity) {
          done();
        });
    });
  }
});
