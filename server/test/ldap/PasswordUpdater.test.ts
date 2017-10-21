import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");
import { PasswordUpdater } from "../../src/lib/ldap/PasswordUpdater";
import { LdapConfiguration } from "../../src/lib/configuration/Configuration";
import { ClientFactoryStub } from "../mocks/ldap/ClientFactoryStub";
import { ClientStub } from "../mocks/ldap/ClientStub";
import { HashGenerator } from "../../src/lib/utils/HashGenerator";

describe("test password update", function () {
  const USERNAME = "username";
  const NEW_PASSWORD = "new-password";

  const ADMIN_USER_DN = "cn=admin,dc=example,dc=com";
  const ADMIN_PASSWORD = "password";

  let clientFactoryStub: ClientFactoryStub;
  let adminClientStub: ClientStub;
  let passwordUpdater: PasswordUpdater;
  let ldapConfig: LdapConfiguration;
  let ssha512HashGenerator: Sinon.SinonStub;

  beforeEach(function () {
    clientFactoryStub = new ClientFactoryStub();
    adminClientStub = new ClientStub();

    ldapConfig = {
      url: "http://ldap",
      user: ADMIN_USER_DN,
      password: ADMIN_PASSWORD,
      users_dn: "ou=users,dc=example,dc=com",
      groups_dn: "ou=groups,dc=example,dc=com",
      group_name_attribute: "cn",
      groups_filter: "cn={0}",
      mail_attribute: "mail",
      users_filter: "cn={0}"
    };

    ssha512HashGenerator = Sinon.stub(HashGenerator, "ssha512");
    passwordUpdater = new PasswordUpdater(ldapConfig, clientFactoryStub);
  });

  afterEach(function () {
    ssha512HashGenerator.restore();
  });

  describe("success", function () {
    it("should update the password successfully", function () {
      clientFactoryStub.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD)
        .returns(adminClientStub);

      ssha512HashGenerator.returns("{CRYPT}$6$abcdefghijklm$AQmxaKfobGY9HSQa6aDYkAWOgPGNhGYn");
      adminClientStub.modifyPasswordStub.withArgs(USERNAME, NEW_PASSWORD).returns(BluebirdPromise.resolve());
      adminClientStub.openStub.returns(BluebirdPromise.resolve());
      adminClientStub.closeStub.returns(BluebirdPromise.resolve());

      return passwordUpdater.updatePassword(USERNAME, NEW_PASSWORD);
    });
  });

  describe("failure", function () {
    it("should fail updating password when modify operation fails", function () {
      clientFactoryStub.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD)
        .returns(adminClientStub);

      ssha512HashGenerator.returns("{CRYPT}$6$abcdefghijklm$AQmxaKfobGY9HSQa6aDYkAWOgPGNhGYn");
      adminClientStub.modifyPasswordStub.withArgs(USERNAME, NEW_PASSWORD)
        .returns(BluebirdPromise.reject(new Error("Error while updating password")));
      adminClientStub.openStub.returns(BluebirdPromise.resolve());
      adminClientStub.closeStub.returns(BluebirdPromise.resolve());

      return passwordUpdater.updatePassword(USERNAME, NEW_PASSWORD)
        .then(function () { return BluebirdPromise.reject(new Error("should not be here")); })
        .catch(function () { return BluebirdPromise.resolve(); });
    });
  });
});