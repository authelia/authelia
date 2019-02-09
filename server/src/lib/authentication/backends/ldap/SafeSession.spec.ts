import BluebirdPromise = require("bluebird");
import { SessionStub } from "./SessionStub.spec";
import { SafeSession } from "./SafeSession";
import Winston = require("winston");

describe("ldap/SanitizedClient", function () {
  let client: SafeSession;

  beforeEach(function () {
    const clientStub = new SessionStub();
    clientStub.searchUserDnStub.onCall(0).returns(BluebirdPromise.resolve());
    clientStub.searchGroupsStub.onCall(0).returns(BluebirdPromise.resolve());
    clientStub.searchEmailsStub.onCall(0).returns(BluebirdPromise.resolve());
    clientStub.modifyPasswordStub.onCall(0).returns(BluebirdPromise.resolve());
    client = new SafeSession(clientStub, Winston);
  });

  describe("special chars are used", function () {
    it("should fail when special chars are used in searchUserDn", function () {
      // potential ldap injection";
      return client.searchUserDn("cn=dummy_user,ou=groupgs")
        .then(function () {
          return BluebirdPromise.reject(new Error("Should not be here."));
        }, function () {
          return BluebirdPromise.resolve();
        });
    });

    it("should fail when special chars are used in searchGroups", function () {
      // potential ldap injection";
      return client.searchGroups("cn=dummy_user,ou=groupgs")
        .then(function () {
          return BluebirdPromise.reject(new Error("Should not be here."));
        }, function () {
          return BluebirdPromise.resolve();
        });
    });

    it("should fail when special chars are used in searchEmails", function () {
      // potential ldap injection";
      return client.searchEmails("cn=dummy_user,ou=groupgs")
        .then(function () {
          return BluebirdPromise.reject(new Error("Should not be here."));
        }, function () {
          return BluebirdPromise.resolve();
        });
    });

    it("should fail when special chars are used in modifyPassword", function () {
      // potential ldap injection";
      return client.modifyPassword("cn=dummy_user,ou=groupgs", "abc")
        .then(function () {
          return BluebirdPromise.reject(new Error("Should not be here."));
        }, function () {
          return BluebirdPromise.resolve();
        });
    });
  });

  describe("no special chars are used", function() {
    it("should succeed when no special chars are used in searchUserDn", function () {
      return client.searchUserDn("dummy_user");
    });

    it("should succeed when no special chars are used in searchGroups", function () {
      return client.searchGroups("dummy_user");
    });

    it("should succeed when no special chars are used in searchEmails", function () {
      return client.searchEmails("dummy_user");
    });

    it("should succeed when no special chars are used in modifyPassword", function () {
      return client.modifyPassword("dummy_user", "abc");
    });
  });
});
