import Sinon = require("sinon");
import Assert = require("assert");
import BluebirdPromise = require("bluebird");

import { Identity } from "../../../../../types/Identity";
import RegistrationHandler from "../../../../../src/lib/routes/secondfactor/u2f/identity/RegistrationHandler";
import ExpressMock = require("../../../../mocks/express");
import { UserDataStoreStub } from "../../../../mocks/storage/UserDataStoreStub";
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../../../mocks/ServerVariablesMockBuilder";
import { ServerVariables } from "../../../../../src/lib/ServerVariables";

describe("test U2F register handler", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    req = ExpressMock.RequestMock();
    req.app = {};
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

    res = ExpressMock.ResponseMock();
    res.send = Sinon.spy();
    res.json = Sinon.spy();
    res.status = Sinon.spy();
  });

  describe("test u2f registration check", test_registration_check);

  function test_registration_check() {
    it("should fail if first_factor has not been passed", function () {
      req.session.auth.first_factor = false;
      return new RegistrationHandler(vars.logger).preValidationInit(req as any)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function (err: Error) {
          return BluebirdPromise.resolve();
        });
    });

    it("should fail if userid is missing", function () {
      req.session.auth.first_factor = false;
      req.session.auth.userid = undefined;

      return new RegistrationHandler(vars.logger).preValidationInit(req as any)
        .then(function () {
          return BluebirdPromise.reject(new Error("should not be here"));
        },
        function (err: Error) {
          return BluebirdPromise.resolve();
        });
    });

    it("should fail if email is missing", function () {
      req.session.auth.first_factor = false;
      req.session.auth.email = undefined;

      return new RegistrationHandler(vars.logger).preValidationInit(req as any)
        .then(function () {
          return BluebirdPromise.reject(new Error("should not be here"));
        },
        function (err: Error) {
          return BluebirdPromise.resolve();
        });
    });

    it("should succeed if first factor passed, userid and email are provided", function () {
      req.session.auth.first_factor = true;
      req.session.auth.email = "admin@example.com";
      req.session.auth.userid = "user";
      return new RegistrationHandler(vars.logger).preValidationInit(req as any);
    });
  }
});
