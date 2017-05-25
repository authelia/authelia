
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
  let ldap_client: LdapjsClientMock;
  let ldapjs: LdapjsMock;
  let ldap_config: LdapConfiguration;

  beforeEach(function () {
    ldap_client = {
      bind: sinon.stub(),
      search: sinon.stub(),
      modify: sinon.stub(),
      on: sinon.stub()
    } as any;

    ldapjs = LdapjsMock();
    ldapjs.createClient.returns(ldap_client);

    ldap_config = {
      url: "http://localhost:324",
      user: "admin",
      password: "password",
      base_dn: "dc=example,dc=com",
      additional_user_dn: "ou=users"
    };

    ldap = new LdapClient.LdapClient(ldap_config, ldapjs, winston);
    return ldap.connect();
  });

  describe("test binding", test_binding);
  describe("test get emails from username", test_get_emails);
  describe("test get groups from username", test_get_groups);
  describe("test update password", test_update_password);

  function test_binding() {
    function test_bind() {
      const username = "username";
      const password = "password";
      return ldap.bind(username, password);
    }

    it("should bind the user if good credentials provided", function () {
      ldap_client.bind.yields();
      return test_bind();
    });

    it("should bind the user with correct DN", function () {
      ldap_config.user_name_attribute = "uid";
      const username = "user";
      const password = "password";
      ldap_client.bind.withArgs("uid=user,ou=users,dc=example,dc=com").yields();
      return ldap.bind(username, password);
    });

    it("should default to cn user search filter if no filter provided", function () {
      const username = "user";
      const password = "password";
      ldap_client.bind.withArgs("cn=user,ou=users,dc=example,dc=com").yields();
      return ldap.bind(username, password);
    });

    it("should not bind the user if wrong credentials provided", function () {
      ldap_client.bind.yields("wrong credentials");
      const promise = test_bind();
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
      ldap_client.search.yields(undefined, res_emitter);

      return ldap.get_emails("user")
        .then(function (emails) {
          assert.deepEqual(emails, [expected_doc.object.mail]);
          return BluebirdPromise.resolve();
        });
    });

    it("should retrieve email for user with uid name attribute", function () {
      ldap_config.user_name_attribute = "uid";
      ldap_client.search.withArgs("uid=username,ou=users,dc=example,dc=com").yields(undefined, res_emitter);
      return ldap.get_emails("username")
        .then(function (emails) {
          assert.deepEqual(emails, ["user@example.com"]);
          return BluebirdPromise.resolve();
        });
    });

    it("should fail on error with search method", function () {
      const expected_doc = {
        mail: ["user@example.com"]
      };
      ldap_client.search.yields("Error while searching mails");

      return ldap.get_emails("user")
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
      ldap_client.search.yields(undefined, res_emitter);
      return ldap.get_groups("user")
        .then(function (groups) {
          assert.deepEqual(groups, ["group1", "group2"]);
          return BluebirdPromise.resolve();
        });
    });

    it("should reduce the scope to additional_group_dn", function (done) {
      ldap_config.additional_group_dn = "ou=groups";
      ldap_client.search.yields(undefined, res_emitter);
      ldap.get_groups("user")
      .then(function() {
        assert.equal(ldap_client.search.getCall(0).args[0], "ou=groups,dc=example,dc=com");
        done();
      });
    });

    it("should use default group_name_attr if not provided", function (done) {
      ldap_client.search.yields(undefined, res_emitter);
      ldap.get_groups("user")
      .then(function() {
        assert.equal(ldap_client.search.getCall(0).args[0], "dc=example,dc=com");
        assert.equal(ldap_client.search.getCall(0).args[1].filter, "member=cn=user,ou=users,dc=example,dc=com");
        assert.deepEqual(ldap_client.search.getCall(0).args[1].attributes, ["cn"]);
        done();
      });
    });

    it("should fail on error with search method", function () {
      ldap_client.search.yields("error");
      return ldap.get_groups("user")
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

      ldap_client.bind.yields(undefined);
      ldap_client.modify.yields(undefined);

      return ldap.update_password("user", "new-password")
        .then(function () {
          assert.deepEqual(ldap_client.modify.getCall(0).args[0], userdn);
          assert.deepEqual(ldap_client.modify.getCall(0).args[1].operation, change.operation);

          const userPassword = ldap_client.modify.getCall(0).args[1].modification.userPassword;
          assert(/{SSHA}/.test(userPassword));
          return BluebirdPromise.resolve();
        });
    });

    it("should fail when ldap throws an error", function () {
      ldap_client.bind.yields(undefined);
      ldap_client.modify.yields("Error");

      return ldap.update_password("user", "new-password")
        .catch(function () {
          return BluebirdPromise.resolve();
        });
    });

    it("should update password of user using particular user name attribute", function () {
      ldap_config.user_name_attribute = "uid";

      ldap_client.bind.yields(undefined);
      ldap_client.modify.withArgs("uid=username,ou=users,dc=example,dc=com").yields();
      return ldap.update_password("username", "newpass");
    });
  }
});

