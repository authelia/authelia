
import { EmailsRetriever } from "../../../../src/server/lib/ldap/EmailsRetriever";
import { LdapConfiguration } from "../../../../src/types/Configuration";

import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import ldapjs = require("ldapjs");
import winston = require("winston");
import { EventEmitter } from "events";

import { LdapjsMock, LdapjsClientMock } from "../mocks/ldapjs";


describe("test emails retriever", function () {
  let emailsRetriever: EmailsRetriever;
  let ldapClient: LdapjsClientMock;
  let ldapjs: LdapjsMock;
  let ldapConfig: LdapConfiguration;
  let adminUserDN: string;
  let adminPassword: string;

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

    emailsRetriever = new EmailsRetriever(ldapConfig, ldapjs, winston);
  });

  function retrieveEmails(ldapClient: LdapjsClientMock) {
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

    const emailsEmitter = {
      on: sinon.spy(function (event: string, fn: (doc: any) => void) {
        if (event != "error") fn(email0);
        if (event != "error") fn(email1);
      })
    };

    ldapClient.search.onCall(0).yields(undefined, emailsEmitter);
  }

  function test_emails_retrieval() {
    const username = "username";
    return emailsRetriever.retrieve(username);
  }

  describe("success", function () {
    beforeEach(function () {
      ldapClient.bind.withArgs(adminUserDN, adminPassword).yields();
      ldapClient.unbind.yields();
    });

    it("should update the password successfully", function () {
      retrieveEmails(ldapClient);
      return test_emails_retrieval();
    });
  });

  describe("failure", function () {
    it("should fail retrieving emails when search operation fails", function () {
      ldapClient.bind.withArgs(adminUserDN, adminPassword).yields();
      ldapClient.search.yields("wrong credentials");
      return test_emails_retrieval()
        .catch(function () {
          return BluebirdPromise.resolve();
        });
    });
  });
});