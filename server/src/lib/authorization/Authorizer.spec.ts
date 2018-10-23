
import Assert = require("assert");
import winston = require("winston");
import { Authorizer } from "./Authorizer";
import { ACLConfiguration, ACLRule } from "../configuration/schema/AclConfiguration";
import { Level } from "./Level";

describe("authorization/Authorizer", function () {
  let authorizer: Authorizer;
  let configuration: ACLConfiguration;

  describe("configuration is null", function() {
    it("should allow access to anything, anywhere for anybody", function() {
      configuration = undefined;
      authorizer = new Authorizer(configuration, winston);

      Assert.equal(authorizer.authorization("home.example.com", "/", "user1", ["group1", "group2"]), Level.BYPASS);
      Assert.equal(authorizer.authorization("home.example.com", "/abc", "user1", ["group1", "group2"]), Level.BYPASS);
      Assert.equal(authorizer.authorization("home.example.com", "/", "user2", ["group1", "group2"]), Level.BYPASS);
      Assert.equal(authorizer.authorization("admin.example.com", "/", "user3", ["group3"]), Level.BYPASS);
    });
  });

  describe("configuration is not null", function () {
    beforeEach(function () {
      configuration = {
        default_policy: "deny",
        rules: []
      };
      authorizer = new Authorizer(configuration, winston);
    });

    describe("check access control with default policy to deny", function () {
      beforeEach(function () {
        configuration.default_policy = "deny";
      });

      it("should deny access when no rule is provided", function () {
        Assert.equal(authorizer.authorization("home.example.com", "/", "user1", ["group1"]), Level.DENY);
      });

      it("should control access when multiple domain matcher is provided", function () {
        configuration.rules = [{
          domain: "*.mail.example.com",
          policy: "two_factor",
          subject: "user:user1",
          resources: [".*"]
        }];
        Assert.equal(authorizer.authorization("home.example.com", "/", "user1", ["group1"]), Level.DENY);
        Assert.equal(authorizer.authorization("mx1.mail.example.com", "/", "user1", ["group1"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("mx1.server.mail.example.com", "/", "user1", ["group1"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("mail.example.com", "/", "user1", ["group1"]), Level.DENY);
      });

      it("should allow access to all resources when resources is not provided", function () {
        configuration.rules = [{
          domain: "*.mail.example.com",
          policy: "two_factor",
          subject: "user:user1"
        }];
        Assert.equal(authorizer.authorization("home.example.com", "/", "user1", ["group1"]), Level.DENY);
        Assert.equal(authorizer.authorization("mx1.mail.example.com", "/", "user1", ["group1"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("mx1.server.mail.example.com", "/", "user1", ["group1"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("mail.example.com", "/", "user1", ["group1"]), Level.DENY);
      });

      describe("check user rules", function () {
        it("should allow access when user has a matching allowing rule", function () {
          configuration.rules = [{
            domain: "home.example.com",
            policy: "two_factor",
            resources: [".*"],
            subject: "user:user1"
          }];
          Assert.equal(authorizer.authorization("home.example.com", "/", "user1", ["group1"]), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization("home.example.com", "/another/resource", "user1", ["group1"]), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization("another.home.example.com", "/", "user1", ["group1"]), Level.DENY);
        });

        it("should deny to other users", function () {
          configuration.rules = [{
            domain: "home.example.com",
            policy: "two_factor",
            resources: [".*"],
            subject: "user:user1"
          }];
          Assert.equal(authorizer.authorization("home.example.com", "/", "user2", ["group1"]), Level.DENY);
          Assert.equal(authorizer.authorization("home.example.com", "/another/resource", "user2", ["group1"]), Level.DENY);
          Assert.equal(authorizer.authorization("another.home.example.com", "/", "user2", ["group1"]), Level.DENY);
        });

        it("should allow user access only to specific resources", function () {
          configuration.rules = [{
            domain: "home.example.com",
            policy: "two_factor",
            resources: ["/private/.*", "^/begin", "/end$"],
            subject: "user:user1"
          }];
          Assert.equal(authorizer.authorization("home.example.com", "/", "user1", ["group1"]), Level.DENY);
          Assert.equal(authorizer.authorization("home.example.com", "/private", "user1", ["group1"]), Level.DENY);
          Assert.equal(authorizer.authorization("home.example.com", "/private/class", "user1", ["group1"]), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization("home.example.com", "/middle/private/class", "user1", ["group1"]), Level.TWO_FACTOR);

          Assert.equal(authorizer.authorization("home.example.com", "/begin", "user1", ["group1"]), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization("home.example.com", "/not/begin", "user1", ["group1"]), Level.DENY);

          Assert.equal(authorizer.authorization("home.example.com", "/abc/end", "user1", ["group1"]), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization("home.example.com", "/abc/end/x", "user1", ["group1"]), Level.DENY);
        });

        it("should allow access to multiple domains", function () {
          configuration.rules = [{
            domain: "home.example.com",
            policy: "two_factor",
            resources: [".*"],
            subject: "user:user1"
          }, {
            domain: "home1.example.com",
            policy: "one_factor",
            resources: [".*"],
            subject: "user:user1"
          }, {
            domain: "home2.example.com",
            policy: "deny",
            resources: [".*"],
            subject: "user:user1"
          }];
          Assert.equal(authorizer.authorization("home.example.com", "/", "user1", ["group1"]), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization("home1.example.com", "/", "user1", ["group1"]), Level.ONE_FACTOR);
          Assert.equal(authorizer.authorization("home2.example.com", "/", "user1", ["group1"]), Level.DENY);
          Assert.equal(authorizer.authorization("home3.example.com", "/", "user1", ["group1"]), Level.DENY);
        });

        it("should apply rules in order", function () {
          configuration.rules = [{
            domain: "home.example.com",
            policy: "one_factor",
            resources: ["/my/private/resource"],
            subject: "user:user1"
          }, {
            domain: "home.example.com",
            policy: "deny",
            resources: ["^/my/private/.*"],
            subject: "user:user1"
          }, {
            domain: "home.example.com",
            policy: "two_factor",
            resources: ["^/my/.*"],
            subject: "user:user1"
          }];

          Assert.equal(authorizer.authorization("home.example.com", "/my/poney", "user1", ["group1"]), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization("home.example.com", "/my/private/duck", "user1", ["group1"]), Level.DENY);
          Assert.equal(authorizer.authorization("home.example.com", "/my/private/resource", "user1", ["group1"]), Level.ONE_FACTOR);
        });
      });

      describe("check group rules", function () {
        it("should allow access when user is in group having a matching allowing rule", function () {
          configuration.rules = [{
            domain: "home.example.com",
            policy: "two_factor",
            resources: ["^/$"],
            subject: "group:group1"
          }, {
            domain: "home.example.com",
            policy: "one_factor",
            resources: ["^/test$"],
            subject: "group:group2"
          }, {
            domain: "home.example.com",
            policy: "deny",
            resources: ["^/private$"],
            subject: "group:group2"
          }];
          Assert.equal(authorizer.authorization("home.example.com", "/", "user1",
            ["group1", "group2", "group3"]), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization("home.example.com", "/test", "user1",
            ["group1", "group2", "group3"]), Level.ONE_FACTOR);
          Assert.equal(authorizer.authorization("home.example.com", "/private", "user1",
            ["group1", "group2", "group3"]), Level.DENY);
          Assert.equal(authorizer.authorization("another.home.example.com", "/", "user1",
            ["group1", "group2", "group3"]), Level.DENY);
        });
      });
    });

    describe("check any rules", function () {
      it("should control access when any rules are defined", function () {
        configuration.rules = [{
          domain: "home.example.com",
          policy: "bypass",
          resources: ["^/public$"]
        }, {
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/private$"]
        }];
        Assert.equal(authorizer.authorization("home.example.com", "/public", "user1",
          ["group1", "group2", "group3"]), Level.BYPASS);
        Assert.equal(authorizer.authorization("home.example.com", "/private", "user1",
          ["group1", "group2", "group3"]), Level.DENY);
        Assert.equal(authorizer.authorization("home.example.com", "/public", "user4",
          ["group5"]), Level.BYPASS);
        Assert.equal(authorizer.authorization("home.example.com", "/private", "user4",
          ["group5"]), Level.DENY);
      });
    });

    describe("check access control with default policy to allow", function () {
      beforeEach(function () {
        configuration.default_policy = "bypass";
      });

      it("should allow access to anything when no rule is provided", function () {
        Assert.equal(authorizer.authorization("home.example.com", "/", "user1", ["group1"]), Level.BYPASS);
        Assert.equal(authorizer.authorization("home.example.com", "/test", "user1", ["group1"]), Level.BYPASS);
        Assert.equal(authorizer.authorization("home.example.com", "/dev", "user1", ["group1"]), Level.BYPASS);
      });

      it("should deny access to one resource when defined", function () {
        configuration.rules = [{
          domain: "home.example.com",
          policy: "deny",
          resources: ["/test"],
          subject: "user:user1"
        }];
        Assert.equal(authorizer.authorization("home.example.com", "/", "user1", ["group1"]), Level.BYPASS);
        Assert.equal(authorizer.authorization("home.example.com", "/test", "user1", ["group1"]), Level.DENY);
        Assert.equal(authorizer.authorization("home.example.com", "/dev", "user1", ["group1"]), Level.BYPASS);
      });
    });

    describe("check access control with complete use case", function () {
      beforeEach(function () {
        configuration.default_policy = "deny";
      });

      it("should control access of multiple user (real use case)", function () {
        // Let say we have three users: admin, john, harry.
        // admin is in groups ["admins"]
        // john is in groups ["dev", "admin-private"]
        // harry is in groups ["dev"]
        configuration.rules = [{
          domain: "home.example.com",
          policy: "two_factor",
          resources: ["^/public$", "^/$"]
        }, {
          domain: "home.example.com",
          policy: "two_factor",
          resources: [".*"],
          subject: "group:admins"
        }, {
          domain: "home.example.com",
          policy: "two_factor",
          resources: ["^/private/?.*"],
          subject: "group:admin-private"
        }, {
          domain: "home.example.com",
          policy: "two_factor",
          resources: ["^/private/john$"],
          subject: "user:john"
        }, {
          domain: "home.example.com",
          policy: "two_factor",
          resources: ["^/private/harry"],
          subject: "user:harry"
        }];

        Assert.equal(authorizer.authorization("home.example.com", "/", "admin", ["admins"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/public", "admin", ["admins"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/dev", "admin", ["admins"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/dev/bob", "admin", ["admins"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/admin", "admin", ["admins"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/private/josh", "admin", ["admins"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/private/john", "admin", ["admins"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/private/harry", "admin", ["admins"]), Level.TWO_FACTOR);

        Assert.equal(authorizer.authorization("home.example.com", "/", "john", ["dev", "admin-private"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/public", "john", ["dev", "admin-private"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/dev", "john", ["dev", "admin-private"]), Level.DENY);
        Assert.equal(authorizer.authorization("home.example.com", "/dev/bob", "john", ["dev", "admin-private"]), Level.DENY);
        Assert.equal(authorizer.authorization("home.example.com", "/admin", "john", ["dev", "admin-private"]), Level.DENY);
        Assert.equal(authorizer.authorization("home.example.com", "/private/josh", "john", ["dev", "admin-private"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/private/john", "john", ["dev", "admin-private"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/private/harry", "john", ["dev", "admin-private"]), Level.TWO_FACTOR);

        Assert.equal(authorizer.authorization("home.example.com", "/", "harry", ["dev"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/public", "harry", ["dev"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/dev", "harry", ["dev"]), Level.DENY);
        Assert.equal(authorizer.authorization("home.example.com", "/dev/bob", "harry", ["dev"]), Level.DENY);
        Assert.equal(authorizer.authorization("home.example.com", "/admin", "harry", ["dev"]), Level.DENY);
        Assert.equal(authorizer.authorization("home.example.com", "/private/josh", "harry", ["dev"]), Level.DENY);
        Assert.equal(authorizer.authorization("home.example.com", "/private/john", "harry", ["dev"]), Level.DENY);
        Assert.equal(authorizer.authorization("home.example.com", "/private/harry", "harry", ["dev"]), Level.TWO_FACTOR);
      });

      it("should allow when allowed at group level and denied at user level", function () {
        configuration.rules = [{
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/dev/bob$"],
          subject: "user:john"
        }, {
          domain: "home.example.com",
          policy: "two_factor",
          resources: ["^/dev/?.*$"],
          subject: "group:dev"
        }];

        Assert.equal(authorizer.authorization("home.example.com", "/dev/john", "john", ["dev"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/dev/bob", "john", ["dev"]), Level.DENY);
      });

      it("should allow access when allowed at 'any' level and denied at user level", function () {
        configuration.rules = [{
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/dev/bob$"],
          subject: "user:john"
        }, {
          domain: "home.example.com",
          policy: "two_factor",
          resources: ["^/dev/?.*$"]
        }];

        Assert.equal(authorizer.authorization("home.example.com", "/dev/john", "john", ["dev"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/dev/bob", "john", ["dev"]), Level.DENY);
      });

      it("should allow access when allowed at 'any' level and denied at group level", function () {
        configuration.rules = [{
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/dev/bob$"],
          subject: "group:dev"
        }, {
          domain: "home.example.com",
          policy: "two_factor",
          resources: ["^/dev/?.*$"]
        }];

        Assert.equal(authorizer.authorization("home.example.com", "/dev/john", "john", ["dev"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/dev/bob", "john", ["dev"]), Level.DENY);
      });

      it("should respect rules precedence", function () {
        // the priority from least to most is 'default_policy', 'all', 'group', 'user'
        // and the first rules in each category as a lower priority than the latest.
        // You can think of it that way: they override themselves inside each category.
        configuration.rules = [{
          domain: "home.example.com",
          policy: "two_factor",
          resources: ["^/dev/?.*$"],
          subject: "user:john"
        }, {
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/dev/bob$"],
          subject: "group:dev"
        }, {
          domain: "home.example.com",
          policy: "two_factor",
          resources: ["^/dev/?.*$"]
        }];

        Assert.equal(authorizer.authorization("home.example.com", "/dev/john", "john", ["dev"]), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization("home.example.com", "/dev/bob", "john", ["dev"]), Level.TWO_FACTOR);
      });
    });
  });
});
