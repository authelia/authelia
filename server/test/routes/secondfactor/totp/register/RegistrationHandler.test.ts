import Sinon = require("sinon");
import winston = require("winston");
import RegistrationHandler from "../../../../../src/lib/routes/secondfactor/totp/identity/RegistrationHandler";
import { Identity } from "../../../../../types/Identity";
import AuthenticationSession = require("../../../../../src/lib/AuthenticationSession");
import { UserDataStore } from "../../../../../src/lib/storage/UserDataStore";
import assert = require("assert");
import BluebirdPromise = require("bluebird");
import ExpressMock = require("../../../../mocks/express");
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../../../mocks/ServerVariablesMockBuilder";
import { ServerVariables } from "../../../../../src/lib/ServerVariables";

describe("test totp register", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    req = ExpressMock.RequestMock();
    req.session = {
      auth: {
        userid: "user",
        email: "user@example.com",
        first_factor: true,
        second_factor: false
      }
    };
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
  });

  describe("test totp registration check", test_registration_check);

  function test_registration_check() {
    it("should fail if first_factor has not been passed", function () {
      req.session.auth.first_factor = false;
      return new RegistrationHandler(vars.logger, vars.userDataStore, vars.totpHandler)
        .preValidationInit(req as any)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function (err: Error) {
          return BluebirdPromise.resolve();
        });
    });

    it("should fail if userid is missing", function (done) {
      req.session.auth.first_factor = false;
      req.session.auth.userid = undefined;

      new RegistrationHandler(vars.logger, vars.userDataStore, vars.totpHandler)
        .preValidationInit(req as any)
        .catch(function (err: Error) {
          done();
        });
    });

    it("should fail if email is missing", function (done) {
      req.session.auth.first_factor = false;
      req.session.auth.email = undefined;

      new RegistrationHandler(vars.logger, vars.userDataStore, vars.totpHandler)
        .preValidationInit(req as any)
        .catch(function (err: Error) {
          done();
        });
    });

    it("should succeed if first factor passed, userid and email are provided", function (done) {
      new RegistrationHandler(vars.logger, vars.userDataStore, vars.totpHandler)
        .preValidationInit(req as any)
        .then(function (identity: Identity) {
          done();
        });
    });
  }
});
