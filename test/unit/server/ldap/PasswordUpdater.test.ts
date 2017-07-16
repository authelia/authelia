
import { PasswordUpdater } from "../../../../src/server/lib/ldap/PasswordUpdater";
import { LdapConfiguration } from "../../../../src/types/Configuration";

import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import ldapjs = require("ldapjs");
import winston = require("winston");
import { EventEmitter } from "events";

import { LdapjsMock, LdapjsClientMock } from "../mocks/ldapjs";


describe("test password update", function () {
  let passwordUpdater: PasswordUpdater;
  let ldapClient: LdapjsClientMock;
  let ldapjs: LdapjsMock;
  let ldapConfig: LdapConfiguration;
  let adminUserDN: string;
  let adminPassword: string;
  let dovehash: any;

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

    dovehash = {
      encode: sinon.stub()
    };

    passwordUpdater = new PasswordUpdater(ldapConfig, ldapjs, dovehash, winston);
  });

  function test_update_password() {
    const username = "username";
    const newpassword = "newpassword";
    return passwordUpdater.updatePassword(username, newpassword);
  }

  describe("success", function () {
    beforeEach(function () {
      ldapClient.bind.withArgs(adminUserDN, adminPassword).yields();
      ldapClient.unbind.yields();
    });

    it("should update the password successfully", function () {
      dovehash.encode.returns("{SSHA}AQmxaKfobGY9HSQa6aDYkAWOgPGNhGYn");
      ldapClient.modify.withArgs("cn=username,ou=users,dc=example,dc=com", {
        operation: "replace",
        modification: {
          userPassword: "{SSHA}AQmxaKfobGY9HSQa6aDYkAWOgPGNhGYn"
        }
      }).yields();
      return test_update_password();
    });
  });

  describe("failure", function () {
    it("should fail updating password when modify operation fails", function () {
      ldapClient.bind.withArgs(adminUserDN, adminPassword).yields();
      ldapClient.modify.yields("wrong credentials");
      return test_update_password()
        .catch(function () {
          return BluebirdPromise.resolve();
        });
    });
  });
});