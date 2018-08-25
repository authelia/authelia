import Assert = require("assert");
import Bluebird = require("bluebird");

import { LdapUsersDatabase } from "./LdapUsersDatabase";

import { SessionFactoryStub } from "./SessionFactoryStub.spec";
import { SessionStub } from "./SessionStub.spec";

const ADMIN_USER_DN = "cn=admin,dc=example,dc=com";
const ADMIN_PASSWORD = "password";

describe("ldap/connector/LdapUsersDatabase", function() {
  let sessionFactory: SessionFactoryStub;
  let usersDatabase: LdapUsersDatabase;

  const USERNAME = "user";
  const PASSWORD = "pass";
  const NEW_PASSWORD = "pass2";

  const LDAP_CONFIG = {
    url: "http://localhost:324",
    additional_users_dn: "ou=users",
    additional_groups_dn: "ou=groups",
    base_dn: "dc=example,dc=com",
    users_filter: "cn={0}",
    groups_filter: "member={0}",
    mail_attribute: "mail",
    group_name_attribute: "cn",
    user: ADMIN_USER_DN,
    password: ADMIN_PASSWORD
  };

  beforeEach(function() {
    sessionFactory = new SessionFactoryStub();
    usersDatabase = new LdapUsersDatabase(sessionFactory, LDAP_CONFIG);
  })

  describe("checkUserPassword", function() {
    it("should return groups and emails when user/password matches", function() {
      const USER_DN = `cn=${USERNAME},dc=example,dc=com`;
      const emails = ["email1", "email2"];
      const groups = ["group1", "group2"];

      const adminSession = new SessionStub();
      const userSession = new SessionStub();

      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(adminSession);
      sessionFactory.createStub.withArgs(USER_DN, PASSWORD).returns(userSession);

      adminSession.openStub.returns(Bluebird.resolve());
      adminSession.closeStub.returns(Bluebird.resolve());
      adminSession.searchUserDnStub.returns(Bluebird.resolve(USER_DN));
      adminSession.searchEmailsStub.withArgs(USERNAME).returns(Bluebird.resolve(emails));
      adminSession.searchGroupsStub.withArgs(USERNAME).returns(Bluebird.resolve(groups));
      
      userSession.openStub.returns(Bluebird.resolve());
      userSession.closeStub.returns(Bluebird.resolve());

      return usersDatabase.checkUserPassword(USERNAME, PASSWORD)
        .then((groupsAndEmails) => {
          Assert.deepEqual(groupsAndEmails.groups, groups);
          Assert.deepEqual(groupsAndEmails.emails, emails);
        })
    });

    it("should fail when username/password is wrong", function() {
      const USER_DN = `cn=${USERNAME},dc=example,dc=com`;

      const adminSession = new SessionStub();
      const userSession = new SessionStub();

      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(adminSession);
      sessionFactory.createStub.withArgs(USER_DN, PASSWORD).returns(userSession);

      adminSession.openStub.returns(Bluebird.resolve());
      adminSession.closeStub.returns(Bluebird.resolve());
      adminSession.searchUserDnStub.returns(Bluebird.resolve(USER_DN));
      
      userSession.openStub.returns(Bluebird.reject(new Error("Failed binding")));
      userSession.closeStub.returns(Bluebird.resolve());

      return usersDatabase.checkUserPassword(USERNAME, PASSWORD)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(userSession.closeStub.called);
          Assert(adminSession.closeStub.called);
          return Bluebird.resolve();
        })
    });

    it("should fail when admin binding fails", function() {
      const USER_DN = `cn=${USERNAME},dc=example,dc=com`;

      const adminSession = new SessionStub();
      const userSession = new SessionStub();

      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(adminSession);
      sessionFactory.createStub.withArgs(USER_DN, PASSWORD).returns(userSession);

      adminSession.openStub.returns(Bluebird.reject(new Error("Failed binding")));
      adminSession.closeStub.returns(Bluebird.resolve());
      adminSession.searchUserDnStub.returns(Bluebird.resolve(USER_DN));

      return usersDatabase.checkUserPassword(USERNAME, PASSWORD)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(userSession.closeStub.notCalled);
          Assert(adminSession.closeStub.called);
          return Bluebird.resolve();
        })
    });

    it("should fail when search for user dn fails", function() {
      const USER_DN = `cn=${USERNAME},dc=example,dc=com`;

      const adminSession = new SessionStub();
      const userSession = new SessionStub();

      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(adminSession);
      sessionFactory.createStub.withArgs(USER_DN, PASSWORD).returns(userSession);

      adminSession.openStub.returns(Bluebird.resolve());
      adminSession.closeStub.returns(Bluebird.resolve());
      adminSession.searchUserDnStub.returns(Bluebird.reject(new Error("Failed searching user dn")));

      return usersDatabase.checkUserPassword(USERNAME, PASSWORD)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(userSession.closeStub.notCalled);
          Assert(adminSession.closeStub.called);
          return Bluebird.resolve();
        })
    });

    it("should fail when groups retrieval fails", function() {
      const USER_DN = `cn=${USERNAME},dc=example,dc=com`;
      const emails = ["email1", "email2"];
      const groups = ["group1", "group2"];

      const adminSession = new SessionStub();
      const userSession = new SessionStub();

      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(adminSession);
      sessionFactory.createStub.withArgs(USER_DN, PASSWORD).returns(userSession);

      adminSession.openStub.returns(Bluebird.resolve());
      adminSession.closeStub.returns(Bluebird.resolve());
      adminSession.searchUserDnStub.returns(Bluebird.resolve(USER_DN));
      adminSession.searchEmailsStub.withArgs(USERNAME)
        .returns(Bluebird.resolve(emails));
      adminSession.searchGroupsStub.withArgs(USERNAME)
        .returns(Bluebird.reject(new Error("Failed retrieving groups")));
      
      userSession.openStub.returns(Bluebird.resolve());
      userSession.closeStub.returns(Bluebird.resolve());

      return usersDatabase.checkUserPassword(USERNAME, PASSWORD)
        .then((groupsAndEmails) => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(userSession.closeStub.called);
          Assert(adminSession.closeStub.called);
        })
    });

    it("should fail when emails retrieval fails", function() {
      const USER_DN = `cn=${USERNAME},dc=example,dc=com`;
      const emails = ["email1", "email2"];
      const groups = ["group1", "group2"];

      const adminSession = new SessionStub();
      const userSession = new SessionStub();

      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(adminSession);
      sessionFactory.createStub.withArgs(USER_DN, PASSWORD).returns(userSession);

      adminSession.openStub.returns(Bluebird.resolve());
      adminSession.closeStub.returns(Bluebird.resolve());
      adminSession.searchUserDnStub.returns(Bluebird.resolve(USER_DN));
      adminSession.searchEmailsStub.withArgs(USERNAME)
        .returns(Bluebird.reject(new Error("Emails retrieval failed")));
      adminSession.searchGroupsStub.withArgs(USERNAME)
        .returns(Bluebird.resolve(groups));
      
      userSession.openStub.returns(Bluebird.resolve());
      userSession.closeStub.returns(Bluebird.resolve());

      return usersDatabase.checkUserPassword(USERNAME, PASSWORD)
        .then((groupsAndEmails) => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(userSession.closeStub.called);
          Assert(adminSession.closeStub.called);
        })
    });
  });

  describe("getEmails", function() {
    it("should succefully retrieves email", () => {
      const emails = ["email1", "email2"];
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.resolve());
      session.closeStub.returns(Bluebird.resolve());
      session.searchEmailsStub.returns(Bluebird.resolve(emails));

      return usersDatabase.getEmails(USERNAME)
        .then((foundEmails) => {
          Assert(session.closeStub.called);
          Assert.deepEqual(foundEmails, emails);
        })
    });

    it("should fail when binding fails", () => {
      const emails = ["email1", "email2"];
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.reject(new Error("Binding failed")));

      return usersDatabase.getEmails(USERNAME)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(session.closeStub.called);
        })
    });

    it("should fail when unbinding fails", () => {
      const emails = ["email1", "email2"];
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.resolve());
      session.searchEmailsStub.returns(Bluebird.resolve(emails));
      session.closeStub.returns(Bluebird.reject(new Error("Unbinding failed")));

      return usersDatabase.getEmails(USERNAME)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(session.closeStub.called);
        })
    });

    it("should fail when search fails", () => {
      const emails = ["email1", "email2"];
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.resolve());
      session.searchEmailsStub.returns(Bluebird.reject(new Error("Search failed")));
      session.closeStub.returns(Bluebird.resolve());

      return usersDatabase.getEmails(USERNAME)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(session.closeStub.called);
        })
    });
  });


  describe("getGroups", function() {
    it("should succefully retrieves groups", () => {
      const groups = ["group1", "group2"];
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.resolve());
      session.closeStub.returns(Bluebird.resolve());
      session.searchGroupsStub.returns(Bluebird.resolve(groups));

      return usersDatabase.getGroups(USERNAME)
        .then((foundGroups) => {
          Assert(session.closeStub.called);
          Assert.deepEqual(foundGroups, groups);
        })
    });

    it("should fail when binding fails", () => {
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.reject(new Error("Binding failed")));

      return usersDatabase.getGroups(USERNAME)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(session.closeStub.called);
        })
    });

    it("should fail when unbinding fails", () => {
      const groups = ["group1", "group2"];
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.resolve());
      session.searchGroupsStub.returns(Bluebird.resolve(groups));
      session.closeStub.returns(Bluebird.reject(new Error("Unbinding failed")));

      return usersDatabase.getGroups(USERNAME)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(session.closeStub.called);
        })
    });

    it("should fail when search fails", () => {
      const groups = ["group1", "group2"];
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.resolve());
      session.searchGroupsStub.returns(Bluebird.reject(new Error("Search failed")));
      session.closeStub.returns(Bluebird.resolve());

      return usersDatabase.getGroups(USERNAME)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch((err) => {
          Assert(session.closeStub.called);
        })
    });
  });


  describe("updatePassword", function() {
    it("should successfully update password", () => {
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.resolve());
      session.closeStub.returns(Bluebird.resolve());
      session.modifyPasswordStub.returns(Bluebird.resolve());

      return usersDatabase.updatePassword(USERNAME, NEW_PASSWORD)
        .then(() => {
          Assert(session.modifyPasswordStub.calledWith(USERNAME, NEW_PASSWORD));
          Assert(session.closeStub.called);
        })
    });

    it("should fail when binding fails", () => {
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.reject(new Error("Binding failed")));
      session.closeStub.returns(Bluebird.resolve());
      session.modifyPasswordStub.returns(Bluebird.resolve());

      return usersDatabase.updatePassword(USERNAME, NEW_PASSWORD)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch(() => {
          Assert(session.closeStub.called);
        })
    });

    it("should fail when update fails", () => {
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.resolve());
      session.closeStub.returns(Bluebird.reject(new Error("Update failed")));
      session.modifyPasswordStub.returns(Bluebird.resolve());

      return usersDatabase.updatePassword(USERNAME, NEW_PASSWORD)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch(() => {
          Assert(session.closeStub.called);
        })
    });

    it("should fail when unbind fails", () => {
      const session = new SessionStub();
      sessionFactory.createStub.withArgs(ADMIN_USER_DN, ADMIN_PASSWORD).returns(session);

      session.openStub.returns(Bluebird.resolve());
      session.closeStub.returns(Bluebird.resolve());
      session.modifyPasswordStub.returns(Bluebird.reject(new Error("Unbind failed")));

      return usersDatabase.updatePassword(USERNAME, NEW_PASSWORD)
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch(() => {
          Assert(session.closeStub.called);
        })
    });
  });
});