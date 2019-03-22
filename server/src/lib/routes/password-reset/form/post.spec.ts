import * as Express from "express";
import PasswordResetFormPost = require("./post");
import { AuthenticationSessionHandler } from "../../../AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../../../types/AuthenticationSession";
import Assert = require("assert");
import BluebirdPromise = require("bluebird");
import ExpressMock = require("../../../stubs/express.spec");
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../../ServerVariables";
import { Level } from "../../../authentication/Level";

describe("routes/password-reset/form/post", function () {
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;
  let vars: ServerVariables;
  let mocks: ServerVariablesMock;
  let authSession: AuthenticationSession;

  beforeEach(function () {
    req = ExpressMock.RequestMock();

    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    mocks.userDataStore.saveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.produceIdentityValidationTokenStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.consumeIdentityValidationTokenStub.returns(BluebirdPromise.resolve({}));

    mocks.config.authentication_backend.ldap = {
      url: "ldap://ldapjs",
      mail_attribute: "mail",
      user: "user",
      password: "password",
      additional_users_dn: "ou=users",
      additional_groups_dn: "ou=groups",
      base_dn: "dc=example,dc=com",
      users_filter: "user",
      group_name_attribute: "cn",
      groups_filter: "groups"
    };

    res = ExpressMock.ResponseMock();
    authSession = AuthenticationSessionHandler.get(req as any, vars.logger);
    authSession.userid = "user";
    authSession.email = "user@example.com";
    authSession.authentication_level = Level.ONE_FACTOR;
  });

  describe("test reset password post", () => {
    it("should update the password and reset auth_session for reauthentication", function () {
      req.body.password = "new-password";

      mocks.usersDatabase.updatePasswordStub.returns(BluebirdPromise.resolve());

      authSession.identity_check = {
        userid: "user",
        challenge: "reset-password"
      };
      return PasswordResetFormPost.default(vars)(req as any, res as any)
        .then(function () {
          return AuthenticationSessionHandler.get(req as any, vars.logger);
        }).then(function (_authSession) {
          Assert.equal(res.status.getCall(0).args[0], 204);
          Assert.equal(_authSession.authentication_level, Level.NOT_AUTHENTICATED);
          return BluebirdPromise.resolve();
        });
    });

    it("should fail if identity_challenge does not exist", function () {
      authSession.identity_check = {
        userid: "user",
        challenge: undefined
      };
      return PasswordResetFormPost.default(vars)(req as any, res as any)
        .then(function () {
          Assert.equal(res.status.getCall(0).args[0], 200);
          Assert.deepEqual(res.send.getCall(0).args[0], {
            error: "An error occurred during password reset. Your password has not been changed."
          });
        });
    });

    it("should fail when ldap fails", function () {
      req.body.password = "new-password";

      mocks.usersDatabase.updatePasswordStub
        .returns(BluebirdPromise.reject("Internal error with LDAP"));

      authSession.identity_check = {
        challenge: "reset-password",
        userid: "user"
      };
      return PasswordResetFormPost.default(vars)(req as any, res as any)
        .then(function () {
          Assert.equal(res.status.getCall(0).args[0], 200);
          Assert.deepEqual(res.send.getCall(0).args[0], {
            error: "An error occurred during password reset. Your password has not been changed."
          });
          return BluebirdPromise.resolve();
        });
    });
  });
});
