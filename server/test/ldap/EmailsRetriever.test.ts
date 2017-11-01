
import { EmailsRetriever } from "../../src/lib/ldap/EmailsRetriever";
import { LdapConfiguration } from "../../src/lib/configuration/Configuration";

import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");

import { ClientFactoryStub } from "../mocks/ldap/ClientFactoryStub";
import { ClientStub } from "../mocks/ldap/ClientStub";

describe("test emails retriever", function () {
  const USERNAME = "username";
  const ADMIN_USER_DN = "cn=admin,dc=example,dc=com";
  const ADMIN_PASSWORD = "password";

  let clientFactoryStub: ClientFactoryStub;
  let adminClientStub: ClientStub;

  let emailsRetriever: EmailsRetriever;
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
      groups_filter: "cn={0}",
      mail_attribute: "mail",
      users_filter: "cn={0}"
    };

    emailsRetriever = new EmailsRetriever(ldapConfig, clientFactoryStub);
  });

  describe("success", function () {
    it("should retrieve emails successfully", function () {
      clientFactoryStub.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD)
        .returns(adminClientStub);

      // admin connects successfully
      adminClientStub.openStub.returns(BluebirdPromise.resolve());
      adminClientStub.closeStub.returns(BluebirdPromise.resolve());

      adminClientStub.searchEmailsStub.withArgs(USERNAME)
        .returns(BluebirdPromise.resolve(["user@example.com"]));

      return emailsRetriever.retrieve(USERNAME);
    });
  });

  describe("failure", function () {
    it("should fail retrieving emails when search operation fails",
      function () {
        clientFactoryStub.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD)
          .returns(adminClientStub);

        // admin connects successfully
        adminClientStub.openStub.returns(BluebirdPromise.resolve());
        adminClientStub.closeStub.returns(BluebirdPromise.resolve());

        adminClientStub.searchEmailsStub.withArgs(USERNAME)
          .rejects(new Error("Error while searching emails"));

        return emailsRetriever.retrieve(USERNAME)
          .then(function () {
            return BluebirdPromise.reject(new Error("Should not be here"));
          })
          .catch(function () {
            return BluebirdPromise.resolve();
          });
      });
  });
});