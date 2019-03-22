import * as Express from "express";
import PasswordResetHandler
  from "./PasswordResetHandler";
import BluebirdPromise = require("bluebird");
import ExpressMock = require("../../../stubs/express.spec");
import { ServerVariablesMock, ServerVariablesMockBuilder }
  from "../../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../../ServerVariables";

describe("routes/password-reset/identity/PasswordResetHandler", function () {
  let req: Express.Request;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.body.username = "user";
    req.session.auth = {
      userid: "user",
      email: "user@example.com",
      first_factor: true,
      second_factor: false
    };
    req.headers.host = "localhost";

    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    mocks.userDataStore.saveU2FRegistrationStub
      .returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.retrieveU2FRegistrationStub
      .returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.produceIdentityValidationTokenStub
      .returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.consumeIdentityValidationTokenStub
      .returns(BluebirdPromise.resolve({}));
  });

  describe("test reset password identity pre check", () => {
    it("should fail when no userid is provided", function () {
      req.body.username = undefined;
      const handler = new PasswordResetHandler(vars.logger,
        vars.usersDatabase);
      return handler.preValidationInit(req as any)
        .then(function () {
          return BluebirdPromise.reject("It should fail");
        })
        .catch(function (err: Error) {
          return BluebirdPromise.resolve();
        });
    });

    it("should fail if ldap fail", function () {
      mocks.usersDatabase.getEmailsStub
        .returns(BluebirdPromise.reject("Internal error"));
      new PasswordResetHandler(vars.logger, vars.usersDatabase)
        .preValidationInit(req as any)
        .then(function () {
          return BluebirdPromise.reject(new Error("should not be here"));
        },
        function (err: Error) {
          return BluebirdPromise.resolve();
        });
    });

    it("should returns identity when ldap replies", function () {
      mocks.usersDatabase.getEmailsStub
        .returns(BluebirdPromise.resolve(["test@example.com"]));
      return new PasswordResetHandler(vars.logger, vars.usersDatabase)
        .preValidationInit(req as any);
    });
  });
});
