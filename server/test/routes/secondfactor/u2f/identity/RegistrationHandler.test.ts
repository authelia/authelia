import sinon = require("sinon");
import winston = require("winston");
import assert = require("assert");
import BluebirdPromise = require("bluebird");

import { Identity } from "../../../../../types/Identity";
import RegistrationHandler from "../../../../../src/lib/routes/secondfactor/u2f/identity/RegistrationHandler";
import AuthenticationSession = require("../../../../../src/lib/AuthenticationSession");

import ExpressMock = require("../../../../mocks/express");
import { UserDataStoreStub } from "../../../../mocks/storage/UserDataStoreStub";
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");

describe("test register handler", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let authSession: AuthenticationSession.AuthenticationSession;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.app = {};
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

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();

    return AuthenticationSession.get(req as any)
      .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
        authSession = _authSession;
        authSession.userid = "user";
        authSession.email = "user@example.com";
        authSession.first_factor = true;
        authSession.second_factor = false;
      });
  });

  describe("test u2f registration check", test_registration_check);

  function test_registration_check() {
    it("should fail if first_factor has not been passed", function () {
      authSession.first_factor = false;
      return new RegistrationHandler().preValidationInit(req as any)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function (err: Error) {
          return BluebirdPromise.resolve();
        });
    });

    it("should fail if userid is missing", function (done) {
      authSession.first_factor = false;
      authSession.userid = undefined;

      new RegistrationHandler().preValidationInit(req as any)
        .catch(function (err: Error) {
          done();
        });
    });

    it("should fail if email is missing", function (done) {
      authSession.first_factor = false;
      authSession.email = undefined;

      new RegistrationHandler().preValidationInit(req as any)
        .catch(function (err: Error) {
          done();
        });
    });

    it("should succeed if first factor passed, userid and email are provided", function (done) {
      new RegistrationHandler().preValidationInit(req as any)
        .then(function (identity: Identity) {
          done();
        });
    });
  }
});
