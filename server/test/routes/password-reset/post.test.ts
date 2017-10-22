
import PasswordResetFormPost = require("../../../src/lib/routes/password-reset/form/post");
import { PasswordUpdater } from "../../../src/lib/ldap/PasswordUpdater";
import { AuthenticationSessionHandler } from "../../../src/lib/AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../types/AuthenticationSession";
import { UserDataStore } from "../../../src/lib/storage/UserDataStore";
import Sinon = require("sinon");
import Assert = require("assert");
import BluebirdPromise = require("bluebird");
import ExpressMock = require("../../mocks/express");
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../mocks/ServerVariablesMockBuilder";
import { ServerVariables } from "../../../src/lib/ServerVariables";

describe("test reset password route", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let vars: ServerVariables;
  let mocks: ServerVariablesMock;
  let authSession: AuthenticationSession;

  beforeEach(function () {
    req = {
      body: {
        userid: "user"
      },
      session: {},
      headers: {
        host: "localhost"
      }
    };

    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    const options = {
      inMemoryOnly: true
    };

    mocks.userDataStore.saveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.produceIdentityValidationTokenStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.consumeIdentityValidationTokenStub.returns(BluebirdPromise.resolve({}));

    mocks.config.ldap = {
      url: "ldap://ldapjs",
      mail_attribute: "mail",
      user: "user",
      password: "password",
      users_dn: "ou=users,dc=example,dc=com",
      groups_dn: "ou=groups,dc=example,dc=com",
      users_filter: "user",
      group_name_attribute: "cn",
      groups_filter: "groups"
    };

    res = ExpressMock.ResponseMock();
    authSession = AuthenticationSessionHandler.get(req as any, vars.logger);
    authSession.userid = "user";
    authSession.email = "user@example.com";
    authSession.first_factor = true;
    authSession.second_factor = false;
  });

  describe("test reset password post", () => {
    it("should update the password and reset auth_session for reauthentication", function () {
      req.body = {};
      req.body.password = "new-password";

      mocks.ldapPasswordUpdater.updatePasswordStub.returns(BluebirdPromise.resolve());

      authSession.identity_check = {
        userid: "user",
        challenge: "reset-password"
      };
      return PasswordResetFormPost.default(vars)(req as any, res as any)
        .then(function () {
          return AuthenticationSessionHandler.get(req as any, vars.logger);
        }).then(function (_authSession) {
          Assert.equal(res.status.getCall(0).args[0], 204);
          Assert.equal(_authSession.first_factor, false);
          Assert.equal(_authSession.second_factor, false);
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
      req.body = {};
      req.body.password = "new-password";

      mocks.ldapPasswordUpdater.updatePasswordStub
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
