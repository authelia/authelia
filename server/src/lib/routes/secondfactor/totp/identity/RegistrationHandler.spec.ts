import RegistrationHandler from "./RegistrationHandler";
import BluebirdPromise = require("bluebird");
import ExpressMock = require("../../../../stubs/express.spec");
import { ServerVariablesMock, ServerVariablesMockBuilder }
  from "../../../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../../../ServerVariables";
import Assert = require("assert");

describe("routes/secondfactor/totp/identity/RegistrationHandler", function () {
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
        second_factor: false,
        identity_check: {
          userid: "user",
          challenge: "totp-register"
        }
      }
    };

    mocks.userDataStore.saveU2FRegistrationStub
      .returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.retrieveU2FRegistrationStub
      .returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.produceIdentityValidationTokenStub
      .returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.consumeIdentityValidationTokenStub
      .returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.saveTOTPSecretStub
      .returns(BluebirdPromise.resolve({}));

    res = ExpressMock.ResponseMock();
  });

  describe("test totp registration pre validation", function () {
    it("should fail if first_factor has not been passed", function () {
      req.session.auth.first_factor = false;
      return new RegistrationHandler(vars.logger, vars.userDataStore,
        vars.totpHandler, vars.config.totp)
        .preValidationInit(req as any)
        .then(function () {
          return BluebirdPromise.reject(new Error("It should fail"));
        })
        .catch(function (err: Error) {
          return BluebirdPromise.resolve();
        });
    });

    it("should fail if userid is missing", function (done) {
      req.session.auth.first_factor = false;
      req.session.auth.userid = undefined;

      new RegistrationHandler(vars.logger, vars.userDataStore, vars.totpHandler,
        vars.config.totp)
        .preValidationInit(req as any)
        .catch(function (err: Error) {
          done();
        });
    });

    it("should fail if email is missing", function (done) {
      req.session.auth.first_factor = false;
      req.session.auth.email = undefined;

      new RegistrationHandler(vars.logger, vars.userDataStore, vars.totpHandler,
        vars.config.totp)
        .preValidationInit(req as any)
        .catch(function (err: Error) {
          done();
        });
    });

    it("should succeed if first factor passed, userid and email are provided",
      function () {
        return new RegistrationHandler(vars.logger, vars.userDataStore,
          vars.totpHandler, vars.config.totp)
          .preValidationInit(req as any);
      });
  });

  describe("test totp registration post validation", function () {
    it("should generate a secret using userId as label and issuer defined in config", function () {
      vars.config.totp = {
        issuer: "issuer"
      };
      return new RegistrationHandler(vars.logger, vars.userDataStore,
        vars.totpHandler, vars.config.totp)
        .postValidationResponse(req as any, res as any)
        .then(function() {
          Assert(mocks.totpHandler.generateStub.calledWithExactly("user", "issuer"));
        });
    });
  });
});
