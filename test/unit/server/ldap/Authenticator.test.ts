
import { Authenticator } from "../../../../src/server/lib/ldap/Authenticator";
import { LdapConfiguration } from "../../../../src/server/lib/configuration/Configuration";

import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");

import { ClientFactoryStub } from "../mocks/ldap/ClientFactoryStub";
import { ClientStub } from "../mocks/ldap/ClientStub";


describe("test ldap authentication", function () {
  const USERNAME = "username";
  const PASSWORD = "password";

  const ADMIN_USER_DN = "cn=admin,dc=example,dc=com";
  const ADMIN_PASSWORD = "admin_password";

  let clientFactoryStub: ClientFactoryStub;
  let adminClientStub: ClientStub;
  let userClientStub: ClientStub;

  let authenticator: Authenticator;
  let ldapConfig: LdapConfiguration;

  beforeEach(function () {
    clientFactoryStub = new ClientFactoryStub();
    adminClientStub = new ClientStub();
    userClientStub = new ClientStub();

    ldapConfig = {
      url: "http://localhost:324",
      users_dn: "ou=users,dc=example,dc=com",
      users_filter: "cn={0}",
      groups_dn: "ou=groups,dc=example,dc=com",
      groups_filter: "member={0}",
      mail_attribute: "mail",
      group_name_attribute: "cn",
      user: ADMIN_USER_DN,
      password: ADMIN_PASSWORD
    };

    authenticator = new Authenticator(ldapConfig, clientFactoryStub);
  });

  describe("success", function () {
    it("should bind the user if good credentials provided", function () {
      clientFactoryStub.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD)
        .returns(adminClientStub);
      clientFactoryStub.createStub.withArgs("cn=" + USERNAME + ",ou=users,dc=example,dc=com", PASSWORD)
        .returns(userClientStub);

      // admin connects successfully
      adminClientStub.openStub.returns(BluebirdPromise.resolve());
      adminClientStub.closeStub.returns(BluebirdPromise.resolve());

      // admin search for user dn of user
      adminClientStub.searchUserDnStub.withArgs(USERNAME)
        .returns(BluebirdPromise.resolve("cn=" + USERNAME + ",ou=users,dc=example,dc=com"));

      // user connects successfully
      userClientStub.openStub.returns(BluebirdPromise.resolve());
      userClientStub.closeStub.returns(BluebirdPromise.resolve());

      // admin retrieves emails and groups of user
      adminClientStub.searchEmailsAndGroupsStub.returns(BluebirdPromise.resolve({
        groups: ["group1"],
        emails: ["user@example.com"]
      }));

      return authenticator.authenticate(USERNAME, PASSWORD);
    });
  });

  describe("failure", function () {
    it("should not bind the user if wrong credentials provided", function () {
      clientFactoryStub.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD)
        .returns(adminClientStub);
      clientFactoryStub.createStub.withArgs("cn=" + USERNAME + ",ou=users,dc=example,dc=com", PASSWORD)
        .returns(userClientStub);

      // admin connects successfully
      adminClientStub.openStub.returns(BluebirdPromise.resolve());
      adminClientStub.closeStub.returns(BluebirdPromise.resolve());

      // admin search for user dn of user
      adminClientStub.searchUserDnStub.withArgs(USERNAME)
        .returns(BluebirdPromise.resolve("cn=" + USERNAME + ",ou=users,dc=example,dc=com"));

      // user connects successfully
      userClientStub.openStub.returns(BluebirdPromise.reject(new Error("Error while binding")));
      userClientStub.closeStub.returns(BluebirdPromise.resolve());

      return authenticator.authenticate(USERNAME, PASSWORD)
        .then(function () { return BluebirdPromise.reject("Should not be here!"); })
        .catch(function () {
          return BluebirdPromise.resolve();
        });
    });

    it("should not bind the user if search of emails or group fails", function () {
      clientFactoryStub.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD)
        .returns(adminClientStub);
      clientFactoryStub.createStub.withArgs("cn=" + USERNAME + ",ou=users,dc=example,dc=com", PASSWORD)
        .returns(userClientStub);

      // admin connects successfully
      adminClientStub.openStub.returns(BluebirdPromise.resolve());
      adminClientStub.closeStub.returns(BluebirdPromise.resolve());

      // admin search for user dn of user
      adminClientStub.searchUserDnStub.withArgs(USERNAME)
        .returns(BluebirdPromise.resolve("cn=" + USERNAME + ",ou=users,dc=example,dc=com"));

      // user connects successfully
      userClientStub.openStub.returns(BluebirdPromise.resolve());
      userClientStub.closeStub.returns(BluebirdPromise.resolve());

      // admin retrieves emails and groups of user
      adminClientStub.searchEmailsAndGroupsStub
        .returns(BluebirdPromise.reject(new Error("Error while retrieving emails and groups")));

      return authenticator.authenticate(USERNAME, PASSWORD)
        .then(function () { return BluebirdPromise.reject("Should not be here!"); })
        .catch(function () {
          return BluebirdPromise.resolve();
        });
    });
  });
});