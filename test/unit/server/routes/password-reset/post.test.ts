
import PasswordResetFormPost = require("../../../../../src/server/lib/routes/password-reset/form/post");
import { PasswordUpdater } from "../../../../../src/server/lib/ldap/PasswordUpdater";
import AuthenticationSession = require("../../../../../src/server/lib/AuthenticationSession");
import { ServerVariables } from "../../../../../src/server/lib/ServerVariables";
import sinon = require("sinon");
import winston = require("winston");
import assert = require("assert");
import BluebirdPromise = require("bluebird");

import ExpressMock = require("../../mocks/express");
import { UserDataStore } from "../../mocks/UserDataStore";
import ServerVariablesMock = require("../../mocks/ServerVariablesMock");

describe("test reset password route", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let userDataStore: UserDataStore;
  let configuration: any;
  let authSession: AuthenticationSession.AuthenticationSession;
  let serverVariables: ServerVariables;

  beforeEach(function () {
    req = {
      body: {
        userid: "user"
      },
      app: {
        get: sinon.stub()
      },
      session: {},
      headers: {
        host: "localhost"
      }
    };

    AuthenticationSession.reset(req as any);
    authSession = AuthenticationSession.get(req as any);
    authSession.userid = "user";
    authSession.email = "user@example.com";
    authSession.first_factor = true;
    authSession.second_factor = false;

    const options = {
      inMemoryOnly: true
    };

    serverVariables = ServerVariablesMock.mock(req.app);
    userDataStore = UserDataStore();
    userDataStore.set_u2f_meta.returns(BluebirdPromise.resolve({}));
    userDataStore.get_u2f_meta.returns(BluebirdPromise.resolve({}));
    userDataStore.issue_identity_check_token.returns(BluebirdPromise.resolve({}));
    userDataStore.consume_identity_check_token.returns(BluebirdPromise.resolve({}));
    serverVariables.userDataStore = userDataStore as any;


    configuration = {
      ldap: {
        base_dn: "dc=example,dc=com",
        user_name_attribute: "cn"
      }
    };

    serverVariables.logger = winston;
    serverVariables.config = configuration;

    serverVariables.ldapPasswordUpdater = {
      updatePassword: sinon.stub()
    } as any;

    res = ExpressMock.ResponseMock();
  });

  describe("test reset password post", () => {
    it("should update the password and reset auth_session for reauthentication", function () {
      authSession.identity_check = {
        userid: "user",
        challenge: "reset-password"
      };
      req.body = {};
      req.body.password = "new-password";

      (serverVariables.ldapPasswordUpdater.updatePassword as sinon.SinonStub).returns(BluebirdPromise.resolve());
      return PasswordResetFormPost.default(req as any, res as any)
        .then(function () {
          const authSession = AuthenticationSession.get(req as any);
          assert.equal(res.status.getCall(0).args[0], 204);
          assert.equal(authSession.first_factor, false);
          assert.equal(authSession.second_factor, false);
          return BluebirdPromise.resolve();
        });
    });

    it("should fail if identity_challenge does not exist", function (done) {
      authSession.identity_check = {
        userid: "user",
        challenge: undefined
      };
      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });
      PasswordResetFormPost.default(req as any, res as any);
    });

    it("should fail when ldap fails", function (done) {
      authSession.identity_check = {
        challenge: "reset-password",
        userid: "user"
      };
      req.body = {};
      req.body.password = "new-password";

      (serverVariables.ldapPasswordUpdater.updatePassword as sinon.SinonStub).returns(BluebirdPromise.reject("Internal error with LDAP"));
      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 500);
        done();
      });
      PasswordResetFormPost.default(req as any, res as any);
    });
  });
});
