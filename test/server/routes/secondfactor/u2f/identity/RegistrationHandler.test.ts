import sinon = require("sinon");
import winston = require("winston");
import assert = require("assert");
import BluebirdPromise = require("bluebird");

import { Identity } from "../../../../../../src/types/Identity";
import RegistrationHandler from "../../../../../../src/server/lib/routes/secondfactor/u2f/identity/RegistrationHandler";
import AuthenticationSession = require("../../../../../../src/server/lib/AuthenticationSession");

import ExpressMock = require("../../../../mocks/express");
import UserDataStoreMock = require("../../../../mocks/UserDataStore");
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");

describe("test register handler", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let userDataStore: UserDataStoreMock.UserDataStore;
  let authSession: AuthenticationSession.AuthenticationSession;

  beforeEach(function () {
    req = ExpressMock.RequestMock;
    req.app = {};
    const mocks = ServerVariablesMock.mock(req.app);
    mocks.logger = winston;
    req.session = {};
    AuthenticationSession.reset(req as any);
    authSession = AuthenticationSession.get(req as any);
    authSession.userid = "user";
    authSession.email = "user@example.com";
    authSession.first_factor = true;
    authSession.second_factor = false;
    req.headers = {};
    req.headers.host = "localhost";

    const options = {
      inMemoryOnly: true
    };

    userDataStore = UserDataStoreMock.UserDataStore();
    userDataStore.set_u2f_meta = sinon.stub().returns(BluebirdPromise.resolve({}));
    userDataStore.get_u2f_meta = sinon.stub().returns(BluebirdPromise.resolve({}));
    userDataStore.issue_identity_check_token = sinon.stub().returns(BluebirdPromise.resolve({}));
    userDataStore.consume_identity_check_token = sinon.stub().returns(BluebirdPromise.resolve({}));
    mocks.userDataStore = userDataStore;

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
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
