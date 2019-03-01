
import { LdapConfiguration } from "../../../configuration/schema/LdapConfiguration";
import { Session } from "./Session";
import { ConnectorFactoryStub } from "./connector/ConnectorFactoryStub.spec";
import { ConnectorStub } from "./connector/ConnectorStub.spec";

import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");
import Winston = require("winston");

describe("ldap/Session", function () {
  const USERNAME = "username";
  const ADMIN_USER_DN = "cn=admin,dc=example,dc=com";
  const ADMIN_PASSWORD = "password";

  it("should replace {0} by username when searching for groups in LDAP", function () {
    const options: LdapConfiguration = {
      url: "ldap://ldap",
      additional_users_dn: "ou=users",
      additional_groups_dn: "ou=groups",
      base_dn: "dc=example,dc=com",
      users_filter: "cn={0}",
      groups_filter: "member=cn={0},ou=users,dc=example,dc=com",
      group_name_attribute: "cn",
      mail_attribute: "mail",
      user: "cn=admin,dc=example,dc=com",
      password: "password"
    };
    const connectorStub = new ConnectorStub();
    connectorStub.searchAsyncStub.returns(BluebirdPromise.resolve([{
      cn: "group1"
    }]));
    const client = new Session(ADMIN_USER_DN, ADMIN_PASSWORD, options, connectorStub, Winston);

    return client.searchGroups("user1")
      .then(function () {
        Assert.equal(connectorStub.searchAsyncStub.getCall(0).args[1].filter,
          "member=cn=user1,ou=users,dc=example,dc=com");
      });
  });

  it("should replace {dn} by user DN when searching for groups in LDAP", function () {
    const USER_DN = "cn=user1,ou=users,dc=example,dc=com";
    const options: LdapConfiguration = {
      url: "ldap://ldap",
      additional_users_dn: "ou=users",
      additional_groups_dn: "ou=groups",
      base_dn: "dc=example,dc=com",
      users_filter: "cn={0}",
      groups_filter: "member={dn}",
      group_name_attribute: "cn",
      mail_attribute: "mail",
      user: "cn=admin,dc=example,dc=com",
      password: "password"
    };
    const ldapClient = new ConnectorStub();

    // Retrieve user DN
    ldapClient.searchAsyncStub.withArgs("ou=users,dc=example,dc=com", {
      scope: "sub",
      sizeLimit: 1,
      attributes: ["dn"],
      filter: "cn=user1"
    }).returns(BluebirdPromise.resolve([{
      dn: USER_DN
    }]));

    // Retrieve groups
    ldapClient.searchAsyncStub.withArgs("ou=groups,dc=example,dc=com", {
      scope: "sub",
      attributes: ["cn"],
      filter: "member=" + USER_DN
    }).returns(BluebirdPromise.resolve([{
      cn: "group1"
    }]));

    const client = new Session(ADMIN_USER_DN, ADMIN_PASSWORD, options, ldapClient, Winston);

    return client.searchGroups("user1")
      .then(function (groups: string[]) {
        Assert.deepEqual(groups, ["group1"]);
      });
  });

  it("should replace {uid} by user uid when searching for groups in LDAP", function () {
    const USER_UID = "user1";
    const options: LdapConfiguration = {
      url: "ldap://ldap",
      additional_users_dn: "ou=users",
      additional_groups_dn: "ou=groups",
      base_dn: "dc=example,dc=com",
      users_filter: "cn={0}",
      groups_filter: "member=cn={uid},ou=users,dc=example,dc=com",
      group_name_attribute: "cn",
      mail_attribute: "mail",
      user: "cn=admin,dc=example,dc=com",
      password: "password"
    };
    const ldapClient = new ConnectorStub();

    // Retrieve user DN
    ldapClient.searchAsyncStub.withArgs("ou=users,dc=example,dc=com", {
      scope: "sub",
      sizeLimit: 1,
      attributes: ["uid"],
      filter: "cn=user1"
    }).returns(BluebirdPromise.resolve([{
      uid: USER_UID
    }]));

    // Retrieve groups
    ldapClient.searchAsyncStub.withArgs("ou=groups,dc=example,dc=com", {
      scope: "sub",
      attributes: ["cn"],
      filter: "member=cn=user1,ou=users,dc=example,dc=com"
    }).returns(BluebirdPromise.resolve([{
      cn: "group1"
    }]));

    const client = new Session(ADMIN_USER_DN, ADMIN_PASSWORD, options, ldapClient, Winston);

    return client.searchGroups("user1")
      .then(function (groups: string[]) {
        Assert.deepEqual(groups, ["group1"]);
      });
  });

  it("should retrieve mail from custom attribute", function () {
    const USER_DN = "cn=user1,ou=users,dc=example,dc=com";
    const options: LdapConfiguration = {
      url: "ldap://ldap",
      additional_users_dn: "ou=users",
      additional_groups_dn: "ou=groups",
      base_dn: "dc=example,dc=com",
      users_filter: "cn={0}",
      groups_filter: "member={dn}",
      group_name_attribute: "cn",
      mail_attribute: "custom_mail",
      user: "cn=admin,dc=example,dc=com",
      password: "password"
    };
    const connector = new ConnectorStub();
    // Retrieve user DN
    connector.searchAsyncStub.withArgs("ou=users,dc=example,dc=com", {
      scope: "sub",
      sizeLimit: 1,
      attributes: ["dn"],
      filter: "cn=user1"
    }).returns(BluebirdPromise.resolve([{
      dn: USER_DN
    }]));

    // Retrieve email
    connector.searchAsyncStub.withArgs("cn=user1,ou=users,dc=example,dc=com", {
      scope: "base",
      sizeLimit: 1,
      attributes: ["custom_mail"],
    }).returns(BluebirdPromise.resolve([{
      custom_mail: "user1@example.com"
    }]));

    const client = new Session(ADMIN_USER_DN, ADMIN_PASSWORD, options, connector, Winston);

    return client.searchEmails("user1")
      .then(function (emails: string[]) {
        Assert.deepEqual(emails, ["user1@example.com"]);
      });
  });
});