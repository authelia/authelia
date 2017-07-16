
import { Authenticator } from "../../../../src/server/lib/ldap/Authenticator";
import { LdapConfiguration } from "../../../../src/types/Configuration";

import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import ldapjs = require("ldapjs");
import winston = require("winston");
import { EventEmitter } from "events";

import { LdapjsMock, LdapjsClientMock } from "../mocks/ldapjs";


describe("test ldap authentication", function () {
  let authenticator: Authenticator;
  let ldapClient: LdapjsClientMock;
  let ldapjs: LdapjsMock;
  let ldapConfig: LdapConfiguration;
  let adminUserDN: string;
  let adminPassword: string;

  function retrieveEmailsAndGroups(ldapClient: LdapjsClientMock) {
    const email0 = {
      object: {
        mail: "user@example.com"
      }
    };

    const email1 = {
      object: {
        mail: "user@example1.com"
      }
    };

    const group0 = {
      object: {
        group: "group0"
      }
    };

    const emailsEmitter = {
      on: sinon.spy(function (event: string, fn: (doc: any) => void) {
        if (event != "error") fn(email0);
        if (event != "error") fn(email1);
      })
    };

    const groupsEmitter = {
      on: sinon.spy(function (event: string, fn: (doc: any) => void) {
        if (event != "error") fn(group0);
      })
    };

    ldapClient.search.onCall(0).yields(undefined, emailsEmitter);
    ldapClient.search.onCall(1).yields(undefined, groupsEmitter);
  }

  beforeEach(function () {
    ldapClient = LdapjsClientMock();
    ldapjs = LdapjsMock();
    ldapjs.createClient.returns(ldapClient);

    // winston.level = "debug";

    adminUserDN = "cn=admin,dc=example,dc=com";
    adminPassword = "password";

    ldapConfig = {
      url: "http://localhost:324",
      user: adminUserDN,
      password: adminPassword,
      base_dn: "dc=example,dc=com",
      additional_user_dn: "ou=users"
    };

    authenticator = new Authenticator(ldapConfig, ldapjs, winston);
  });

  function test_check_password_internal() {
    const username = "username";
    const password = "password";
    return authenticator.authenticate(username, password);
  }

  describe("success", function () {
    beforeEach(function () {
      retrieveEmailsAndGroups(ldapClient);
      ldapClient.bind.withArgs(adminUserDN, adminPassword).yields();
      ldapClient.unbind.yields();
    });

    it("should bind the user if good credentials provided", function () {
      ldapClient.bind.withArgs("cn=username,ou=users,dc=example,dc=com", "password").yields();
      return test_check_password_internal();
    });

    it("should bind the user with correct DN", function () {
      ldapConfig.user_name_attribute = "uid";
      ldapClient.bind.withArgs("uid=username,ou=users,dc=example,dc=com", "password").yields();
      return test_check_password_internal();
    });
  });

  describe("failure", function () {
    it("should not bind the user if wrong credentials provided", function () {
      ldapClient.bind.yields("wrong credentials");
      return test_check_password_internal()
        .catch(function () {
          return BluebirdPromise.resolve();
        });
    });

    it("should not bind the user if search of emails or group fails", function () {
      ldapClient.bind.withArgs("cn=username,ou=users,dc=example,dc=com", "password").yields();
      ldapClient.bind.withArgs(adminUserDN, adminPassword).yields();
      ldapClient.unbind.yields();
      ldapClient.search.yields("wrong credentials");
      return test_check_password_internal()
        .catch(function () {
          return BluebirdPromise.resolve();
        });
    });
  });
});