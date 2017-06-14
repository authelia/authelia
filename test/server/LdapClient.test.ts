
import LdapClient = require("../../src/server/lib/LdapClient");
import { LdapConfiguration } from "../../src/types/Configuration";

import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import ldapjs = require("ldapjs");
import winston = require("winston");
import { EventEmitter } from "events";

import { LdapjsMock, LdapjsClientMock } from "./mocks/ldapjs";


describe("test ldap validation", function () {
  let ldap: LdapClient.LdapClient;
  let ldapClient: LdapjsClientMock;
  let ldapjs: LdapjsMock;
  let ldapConfig: LdapConfiguration;

  beforeEach(function () {
    ldapClient = LdapjsClientMock();
    ldapjs = LdapjsMock();
    ldapjs.createClient.returns(ldapClient);

    ldapConfig = {
      url: "http://localhost:324",
      user: "admin",
      password: "password",
      base_dn: "dc=example,dc=com",
      additional_user_dn: "ou=users"
    };

    ldap = new LdapClient.LdapClient(ldapConfig, ldapjs, winston);
  });

  describe("test checking password", test_checking_password);
  describe("test get emails from username", test_get_emails);
  describe("test get groups from username", test_get_groups);
  describe("test update password", test_update_password);

  function test_checking_password() {
    function test_check_password_internal() {
      const username = "username";
      const password = "password";
      return ldap.checkPassword(username, password);
    }

    it("should bind the user if good credentials provided", function () {
      ldapClient.bind.yields();
      ldapClient.unbind.yields();
      return test_check_password_internal();
    });

    it("should bind the user with correct DN", function () {
      ldapConfig.user_name_attribute = "uid";
      const username = "user";
      const password = "password";
      ldapClient.bind.withArgs("uid=user,ou=users,dc=example,dc=com").yields();
      ldapClient.unbind.yields();
      return ldap.checkPassword(username, password);
    });

    it("should default to cn user search filter if no filter provided", function () {
      const username = "user";
      const password = "password";
      ldapClient.bind.withArgs("cn=user,ou=users,dc=example,dc=com").yields();
      ldapClient.unbind.yields();
      return ldap.checkPassword(username, password);
    });

    it("should not bind the user if wrong credentials provided", function () {
      ldapClient.bind.yields("wrong credentials");
      const promise = test_check_password_internal();
      return promise.catch(function () {
        return BluebirdPromise.resolve();
      });
    });
  }

  function test_get_emails() {
    let res_emitter: any;
    let expected_doc: any;

    beforeEach(function () {
      expected_doc = {
        object: {
          mail: "user@example.com"
        }
      };

      res_emitter = {
        on: sinon.spy(function (event: string, fn: (doc: any) => void) {
          if (event != "error") fn(expected_doc);
        })
      };
    });

    it("should retrieve the email of an existing user", function () {
      ldapClient.search.yields(undefined, res_emitter);

      return ldap.retrieveEmails("user")
        .then(function (emails) {
          assert.deepEqual(emails, [expected_doc.object.mail]);
          return BluebirdPromise.resolve();
        });
    });

    it("should retrieve email for user with uid name attribute", function () {
      ldapConfig.user_name_attribute = "uid";
      ldapClient.search.withArgs("uid=username,ou=users,dc=example,dc=com").yields(undefined, res_emitter);
      return ldap.retrieveEmails("username")
        .then(function (emails) {
          assert.deepEqual(emails, ["user@example.com"]);
          return BluebirdPromise.resolve();
        });
    });

    it("should fail on error with search method", function () {
      const expected_doc = {
        mail: ["user@example.com"]
      };
      ldapClient.search.yields("Error while searching mails");

      return ldap.retrieveEmails("user")
        .catch(function () {
          return BluebirdPromise.resolve();
        });
    });
  }

  function test_get_groups() {
    let res_emitter: any;
    let expected_doc1: any, expected_doc2: any;

    beforeEach(function () {
      expected_doc1 = {
        object: {
          cn: "group1"
        }
      };

      expected_doc2 = {
        object: {
          cn: "group2"
        }
      };

      res_emitter = {
        on: sinon.spy(function (event: string, fn: (doc: any) => void) {
          if (event != "error") fn(expected_doc1);
          if (event != "error") fn(expected_doc2);
        })
      };
    });

    it("should retrieve the groups of an existing user", function () {
      ldapClient.search.yields(undefined, res_emitter);
      return ldap.retrieveGroups("user")
        .then(function (groups) {
          assert.deepEqual(groups, ["group1", "group2"]);
          return BluebirdPromise.resolve();
        });
    });

    it("should reduce the scope to additional_group_dn", function (done) {
      ldapConfig.additional_group_dn = "ou=groups";
      ldapClient.search.yields(undefined, res_emitter);
      ldap.retrieveGroups("user")
      .then(function() {
        assert.equal(ldapClient.search.getCall(0).args[0], "ou=groups,dc=example,dc=com");
        done();
      });
    });

    it("should use default group_name_attr if not provided", function (done) {
      ldapClient.search.yields(undefined, res_emitter);
      ldap.retrieveGroups("user")
      .then(function() {
        assert.equal(ldapClient.search.getCall(0).args[0], "dc=example,dc=com");
        assert.equal(ldapClient.search.getCall(0).args[1].filter, "member=cn=user,ou=users,dc=example,dc=com");
        assert.deepEqual(ldapClient.search.getCall(0).args[1].attributes, ["cn"]);
        done();
      });
    });

    it("should fail on error with search method", function () {
      ldapClient.search.yields("error");
      return ldap.retrieveGroups("user")
        .catch(function () {
          return BluebirdPromise.resolve();
        });
    });
  }

  function test_update_password() {
    it("should update the password successfully", function () {
      const change = {
        operation: "replace",
        modification: {
          userPassword: "new-password"
        }
      };
      const userdn = "cn=user,ou=users,dc=example,dc=com";

      ldapClient.bind.yields();
      ldapClient.unbind.yields();
      ldapClient.modify.yields();

      return ldap.updatePassword("user", "new-password")
        .then(function () {
          assert.deepEqual(ldapClient.modify.getCall(0).args[0], userdn);
          assert.deepEqual(ldapClient.modify.getCall(0).args[1].operation, change.operation);

          const userPassword = ldapClient.modify.getCall(0).args[1].modification.userPassword;
          assert(/{SSHA}/.test(userPassword));
          return BluebirdPromise.resolve();
        })
        .catch(function(err) { return BluebirdPromise.reject(new Error("It should fail")); });
    });

    it("should fail when ldap throws an error", function () {
      ldapClient.bind.yields(undefined);
      ldapClient.modify.yields("Error");

      return ldap.updatePassword("user", "new-password")
        .catch(function () {
          return BluebirdPromise.resolve();
        });
    });

    it("should update password of user using particular user name attribute", function () {
      ldapConfig.user_name_attribute = "uid";

      ldapClient.bind.yields();
      ldapClient.unbind.yields();
      ldapClient.modify.withArgs("uid=username,ou=users,dc=example,dc=com").yields();
      return ldap.updatePassword("username", "newpass");
    });
  }
});

