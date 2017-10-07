
import PasswordResetFormPost = require("../../../src/lib/routes/password-reset/form/post");
import { PasswordUpdater } from "../../../src/lib/ldap/PasswordUpdater";
import AuthenticationSession = require("../../../src/lib/AuthenticationSession");
import { ServerVariablesHandler } from "../../../src/lib/ServerVariablesHandler";
import { UserDataStore } from "../../../src/lib/storage/UserDataStore";
import Sinon = require("sinon");
import winston = require("winston");
import assert = require("assert");
import BluebirdPromise = require("bluebird");

import ExpressMock = require("../../mocks/express");
import ServerVariablesMock = require("../../mocks/ServerVariablesMock");

describe("test reset password route", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let configuration: any;
  let serverVariables: ServerVariablesMock.ServerVariablesMock;

  beforeEach(function () {
    req = {
      body: {
        userid: "user"
      },
      app: {
        get: Sinon.stub().returns({ logger: winston })
      },
      session: {},
      headers: {
        host: "localhost"
      }
    };

    AuthenticationSession.reset(req as any);

    const options = {
      inMemoryOnly: true
    };

    serverVariables = ServerVariablesMock.mock(req.app);
    serverVariables.userDataStore.saveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));
    serverVariables.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));
    serverVariables.userDataStore.produceIdentityValidationTokenStub.returns(BluebirdPromise.resolve({}));
    serverVariables.userDataStore.consumeIdentityValidationTokenStub.returns(BluebirdPromise.resolve({}));

    configuration = {
      ldap: {
        base_dn: "dc=example,dc=com",
        user_name_attribute: "cn"
      }
    };

    serverVariables.logger = winston;
    serverVariables.config = configuration;

    serverVariables.ldapPasswordUpdater = {
      updatePassword: Sinon.stub()
    } as any;

    res = ExpressMock.ResponseMock();
    AuthenticationSession.get(req as any)
      .then(function (authSession: AuthenticationSession.AuthenticationSession) {
        authSession.userid = "user";
        authSession.email = "user@example.com";
        authSession.first_factor = true;
        authSession.second_factor = false;
      });
  });

  describe("test reset password post", () => {
    it("should update the password and reset auth_session for reauthentication", function () {
      req.body = {};
      req.body.password = "new-password";

      (serverVariables.ldapPasswordUpdater.updatePassword as sinon.SinonStub).returns(BluebirdPromise.resolve());

      return AuthenticationSession.get(req as any)
        .then(function (authSession) {
          authSession.identity_check = {
            userid: "user",
            challenge: "reset-password"
          };
          return PasswordResetFormPost.default(req as any, res as any);
        })
        .then(function () {
          return AuthenticationSession.get(req as any);
        }).then(function (_authSession: AuthenticationSession.AuthenticationSession) {
          assert.equal(res.status.getCall(0).args[0], 204);
          assert.equal(_authSession.first_factor, false);
          assert.equal(_authSession.second_factor, false);
          return BluebirdPromise.resolve();
        });
    });

    it("should fail if identity_challenge does not exist", function () {
      return AuthenticationSession.get(req as any)
        .then(function (authSession) {
          authSession.identity_check = {
            userid: "user",
            challenge: undefined
          };
          return PasswordResetFormPost.default(req as any, res as any);
        })
        .then(function () {
          assert.equal(res.status.getCall(0).args[0], 403);
        });
    });

    it("should fail when ldap fails", function () {
      req.body = {};
      req.body.password = "new-password";

      (serverVariables.ldapPasswordUpdater.updatePassword as Sinon.SinonStub)
        .returns(BluebirdPromise.reject("Internal error with LDAP"));

      return AuthenticationSession.get(req as any)
        .then(function (authSession) {
          authSession.identity_check = {
            challenge: "reset-password",
            userid: "user"
          };
          return PasswordResetFormPost.default(req as any, res as any);
        }).then(function () {
          assert.equal(res.status.getCall(0).args[0], 500);
          return BluebirdPromise.resolve();
        });
    });
  });
});
