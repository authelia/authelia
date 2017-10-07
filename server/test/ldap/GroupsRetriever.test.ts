
import { GroupsRetriever } from "../../src/lib/ldap/GroupsRetriever";
import { LdapConfiguration } from "../../src/lib/configuration/Configuration";

import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");

import { ClientFactoryStub } from "../mocks/ldap/ClientFactoryStub";
import { ClientStub } from "../mocks/ldap/ClientStub";

describe("test groups retriever", function () {
  const USERNAME = "username";
  const ADMIN_USER_DN = "cn=admin,dc=example,dc=com";
  const ADMIN_PASSWORD = "password";

  let clientFactoryStub: ClientFactoryStub;
  let adminClientStub: ClientStub;

  let groupsRetriever: GroupsRetriever;
  let ldapConfig: LdapConfiguration;

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
      groups_filter: "member=cn={0},ou=users,dc=example,dc=com",
      mail_attribute: "mail",
      users_filter: "cn={0}"
    };

    groupsRetriever = new GroupsRetriever(ldapConfig, clientFactoryStub);
  });

  describe("success", function () {
    it("should retrieve groups successfully", function () {
      clientFactoryStub.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD)
        .returns(adminClientStub);

      // admin connects successfully
      adminClientStub.openStub.returns(BluebirdPromise.resolve());
      adminClientStub.closeStub.returns(BluebirdPromise.resolve());

      adminClientStub.searchGroupsStub.withArgs(USERNAME)
        .returns(BluebirdPromise.resolve(["user@example.com"]));

      return groupsRetriever.retrieve(USERNAME);
    });
  });

  describe("failure", function () {
    it("should fail retrieving groups when search operation fails", function () {
      clientFactoryStub.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD)
        .returns(adminClientStub);

      // admin connects successfully
      adminClientStub.openStub.returns(BluebirdPromise.resolve());
      adminClientStub.closeStub.returns(BluebirdPromise.resolve());

      adminClientStub.searchGroupsStub.withArgs(USERNAME)
        .returns(BluebirdPromise.reject(new Error("Error while searching groups")));

      return groupsRetriever.retrieve(USERNAME)
        .then(function () { return BluebirdPromise.reject(new Error("Should not be here")); })
        .catch(function () { return BluebirdPromise.resolve(); });
    });
  });
});