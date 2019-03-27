
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

      Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user1", groups: ["group1", "group2"]}, "127.0.0.1"), Level.BYPASS);
      Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/abc"}, {user: "user1", groups: ["group1", "group2"]}, "127.0.0.1"), Level.BYPASS);
      Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user2", groups: ["group1", "group2"]}, "127.0.0.1"), Level.BYPASS);
      Assert.equal(authorizer.authorization({domain: "admin.example.com", resource: "/"}, {user: "user3", groups: ["group3"]}, "127.0.0.1"), Level.BYPASS);
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
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
      });

      it("should control access when multiple domain matcher is provided", function () {
        configuration.rules = [{
          domain: "*.mail.example.com",
          policy: "two_factor",
          subject: "user:user1",
          resources: [".*"]
        }];
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "mx1.mail.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "mx1.server.mail.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "mail.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
      });

      it("should allow access to all resources when resources is not provided", function () {
        configuration.rules = [{
          domain: "*.mail.example.com",
          policy: "two_factor",
          subject: "user:user1"
        }];
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "mx1.mail.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "mx1.server.mail.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "mail.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
      });

      describe("check user rules", function () {
        it("should allow access when user has a matching allowing rule", function () {
          configuration.rules = [{
            domain: "home.example.com",
            policy: "two_factor",
            resources: [".*"],
            subject: "user:user1"
          }];
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/another/resource"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization({domain: "another.home.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
        });

        it("should deny to other users", function () {
          configuration.rules = [{
            domain: "home.example.com",
            policy: "two_factor",
            resources: [".*"],
            subject: "user:user1"
          }];
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user2", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/another/resource"}, {user: "user2", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
          Assert.equal(authorizer.authorization({domain: "another.home.example.com", resource: "/"}, {user: "user2", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
        });

        it("should allow user access only to specific resources", function () {
          configuration.rules = [{
            domain: "home.example.com",
            policy: "two_factor",
            resources: ["/private/.*", "^/begin", "/end$"],
            subject: "user:user1"
          }];
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private/class"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/middle/private/class"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);

          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/begin"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/not/begin"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);

          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/abc/end"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/abc/end/x"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
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
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization({domain: "home1.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.ONE_FACTOR);
          Assert.equal(authorizer.authorization({domain: "home2.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
          Assert.equal(authorizer.authorization({domain: "home3.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
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

          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/my/poney"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/my/private/duck"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/my/private/resource"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.ONE_FACTOR);
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
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"},
            {user: "user1", groups: ["group1", "group2", "group3"]}, "127.0.0.1"), Level.TWO_FACTOR);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/test"},
            {user: "user1", groups: ["group1", "group2", "group3"]}, "127.0.0.1"), Level.ONE_FACTOR);
          Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private"},
            {user: "user1", groups: ["group1", "group2", "group3"]}, "127.0.0.1"), Level.DENY);
          Assert.equal(authorizer.authorization({domain: "another.home.example.com", resource: "/"},
            {user: "user1", groups: ["group1", "group2", "group3"]}, "127.0.0.1"), Level.DENY);
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
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/public"},
          {user: "user1", groups: ["group1", "group2", "group3"]}, "127.0.0.1"), Level.BYPASS);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private"},
          {user: "user1", groups: ["group1", "group2", "group3"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/public"},
          {user: "user4", groups: ["group5"]}, "127.0.0.1"), Level.BYPASS);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private"},
          {user: "user4", groups: ["group5"]}, "127.0.0.1"), Level.DENY);
      });
    });

    describe("check access control with default policy to allow", function () {
      beforeEach(function () {
        configuration.default_policy = "bypass";
      });

      it("should allow access to anything when no rule is provided", function () {
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.BYPASS);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/test"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.BYPASS);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.BYPASS);
      });

      it("should deny access to one resource when defined", function () {
        configuration.rules = [{
          domain: "home.example.com",
          policy: "deny",
          resources: ["/test"],
          subject: "user:user1"
        }];
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.BYPASS);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/test"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev"}, {user: "user1", groups: ["group1"]}, "127.0.0.1"), Level.BYPASS);
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

        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "admin", groups: ["admins"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/public"}, {user: "admin", groups: ["admins"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev"}, {user: "admin", groups: ["admins"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/bob"}, {user: "admin", groups: ["admins"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/admin"}, {user: "admin", groups: ["admins"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private/josh"}, {user: "admin", groups: ["admins"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private/john"}, {user: "admin", groups: ["admins"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private/harry"}, {user: "admin", groups: ["admins"]}, "127.0.0.1"), Level.TWO_FACTOR);

        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "john", groups: ["dev", "admin-private"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/public"}, {user: "john", groups: ["dev", "admin-private"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev"}, {user: "john", groups: ["dev", "admin-private"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/bob"}, {user: "john", groups: ["dev", "admin-private"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/admin"}, {user: "john", groups: ["dev", "admin-private"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private/josh"}, {user: "john", groups: ["dev", "admin-private"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private/john"}, {user: "john", groups: ["dev", "admin-private"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private/harry"}, {user: "john", groups: ["dev", "admin-private"]}, "127.0.0.1"), Level.TWO_FACTOR);

        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/"}, {user: "harry", groups: ["dev"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/public"}, {user: "harry", groups: ["dev"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev"}, {user: "harry", groups: ["dev"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/bob"}, {user: "harry", groups: ["dev"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/admin"}, {user: "harry", groups: ["dev"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private/josh"}, {user: "harry", groups: ["dev"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private/john"}, {user: "harry", groups: ["dev"]}, "127.0.0.1"), Level.DENY);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/private/harry"}, {user: "harry", groups: ["dev"]}, "127.0.0.1"), Level.TWO_FACTOR);
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

        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/john"}, {user: "john", groups: ["dev"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/bob"}, {user: "john", groups: ["dev"]}, "127.0.0.1"), Level.DENY);
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

        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/john"}, {user: "john", groups: ["dev"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/bob"}, {user: "john", groups: ["dev"]}, "127.0.0.1"), Level.DENY);
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

        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/john"}, {user: "john", groups: ["dev"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/bob"}, {user: "john", groups: ["dev"]}, "127.0.0.1"), Level.DENY);
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

        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/john"}, {user: "john", groups: ["dev"]}, "127.0.0.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/bob"}, {user: "john", groups: ["dev"]}, "127.0.0.1"), Level.TWO_FACTOR);
      });
    });

    describe("check network rules", function () {
      beforeEach(function () {
        configuration.rules = [{
          domain: "home.example.com",
          policy: "one_factor",
          subject: "user:john",
          networks: ["192.168.0.0/24", "10.0.0.0/8"]
        },
        {
          domain: "home.example.com",
          policy: "two_factor",
          subject: "user:john",
        },
        {
          domain: "public.example.com",
          policy: "bypass",
          networks: ["10.0.0.0/8"]
        }];
      });

      it("should respect network ranges", function() {
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/john"}, {user: "john", groups: ["dev"]}, "192.168.4.1"), Level.TWO_FACTOR);
        Assert.equal(authorizer.authorization({domain: "home.example.com", resource: "/dev/bob"}, {user: "john", groups: ["dev"]}, "192.168.0.5"), Level.ONE_FACTOR);
        Assert.equal(authorizer.authorization({domain: "public.example.com", resource: "/dev/bob"}, {user: "john", groups: ["dev"]}, "10.1.3.0"), Level.BYPASS);
        Assert.equal(authorizer.authorization({domain: "public.example.com", resource: "/dev/bob"}, {user: "john", groups: ["dev"]}, "11.1.3.0"), Level.DENY);
      });
    });
  });
});
