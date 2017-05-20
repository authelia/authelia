
import reset_password = require("../../../src/lib/routes/reset_password");
import LdapClient = require("../../../src/lib/LdapClient");
import sinon = require("sinon");
import winston = require("winston");
import assert = require("assert");
import BluebirdPromise = require("bluebird");

import ExpressMock = require("../mocks/express");
import { LdapClientMock } from "../mocks/LdapClient";
import { UserDataStore } from "../mocks/UserDataStore";

describe("test reset password", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let user_data_store: UserDataStore;
  let ldap_client: LdapClientMock;
  let configuration: any;

  beforeEach(function () {
    req = {
      body: {
        userid: "user"
      },
      app: {
        get: sinon.stub()
      },
      session: {
        auth_session: {
          userid: "user",
          email: "user@example.com",
          first_factor: true,
          second_factor: false
        }
      },
      headers: {
        host: "localhost"
      }
    };

    const options = {
      inMemoryOnly: true
    };

    user_data_store = UserDataStore();
    user_data_store.set_u2f_meta.returns(Promise.resolve({}));
    user_data_store.get_u2f_meta.returns(Promise.resolve({}));
    user_data_store.issue_identity_check_token.returns(Promise.resolve({}));
    user_data_store.consume_identity_check_token.returns(Promise.resolve({}));
    req.app.get.withArgs("user data store").returns(user_data_store);


    configuration = {
      ldap: {
        base_dn: "dc=example,dc=com",
        user_name_attribute: "cn"
      }
    };

    req.app.get.withArgs("logger").returns(winston);
    req.app.get.withArgs("config").returns(configuration);

    ldap_client = LdapClientMock();
    req.app.get.withArgs("ldap").returns(ldap_client);

    res = ExpressMock.ResponseMock();
  });

  describe("test reset password identity pre check", test_reset_password_check);
  describe("test reset password post", test_reset_password_post);

  function test_reset_password_check() {
    it("should fail when no userid is provided", function (done) {
      req.body.userid = undefined;
      reset_password.icheck_interface.pre_check_callback(req)
        .catch(function (err: Error) {
          done();
        });
    });

    it("should fail if ldap fail", function (done) {
      ldap_client.get_emails.returns(BluebirdPromise.reject("Internal error"));
      reset_password.icheck_interface.pre_check_callback(req)
        .catch(function (err: Error) {
          done();
        });
    });

    it("should perform a search in ldap to find email address", function (done) {
      configuration.ldap.user_name_attribute = "uid";
      ldap_client.get_emails.returns(BluebirdPromise.resolve([]));
      reset_password.icheck_interface.pre_check_callback(req)
        .then(function () {
          assert.equal("user", ldap_client.get_emails.getCall(0).args[0]);
          done();
        });
    });

    it("should returns identity when ldap replies", function (done) {
      ldap_client.get_emails.returns(BluebirdPromise.resolve(["test@example.com"]));
      reset_password.icheck_interface.pre_check_callback(req)
        .then(function () {
          done();
        });
    });
  }

  function test_reset_password_post() {
    it("should update the password and reset auth_session for reauthentication", function (done) {
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.userid = "user";
      req.session.auth_session.identity_check.challenge = "reset-password";
      req.body = {};
      req.body.password = "new-password";

      ldap_client.update_password.returns(BluebirdPromise.resolve());
      ldap_client.bind.returns(BluebirdPromise.resolve());
      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 204);
        assert.equal(req.session.auth_session, undefined);
        done();
      });
      reset_password.post(req, res);
    });

    it("should fail if identity_challenge does not exist", function (done) {
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.challenge = undefined;
      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });
      reset_password.post(req, res);
    });

    it("should fail when ldap fails", function (done) {
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.challenge = "reset-password";
      req.body = {};
      req.body.password = "new-password";

      ldap_client.bind.yields(undefined);
      ldap_client.update_password.returns(BluebirdPromise.reject("Internal error with LDAP"));
      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 500);
        done();
      });
      reset_password.post(req, res);
    });
  }
});
