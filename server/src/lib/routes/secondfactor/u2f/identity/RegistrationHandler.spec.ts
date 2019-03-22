import * as Express from "express";
import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import RegistrationHandler from "./RegistrationHandler";
import ExpressMock = require("../../../../stubs/express.spec");
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../../../ServerVariables";

describe("routes/secondfactor/u2f/identity/RegistrationHandler", function () {
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    req = ExpressMock.RequestMock();
    req.session = {
      ...req.session,
      auth: {
        userid: "user",
        email: "user@example.com",
        first_factor: true,
        second_factor: false
      }
    };
    req.headers = {};
    req.headers.host = "localhost";

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
