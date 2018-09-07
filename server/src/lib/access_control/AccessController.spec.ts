import Assert = require("assert");
import winston = require("winston");
import { AccessController } from "./AccessController";
import { ACLConfiguration } from "../configuration/schema/AclConfiguration";
import { WhitelistValue } from "../authentication/whitelist/WhitelistHandler";

describe("access_control/AccessController", function () {
  let accessController: AccessController;
  let configuration: ACLConfiguration;

  describe("configuration is null", function() {
    it("should allow access to anything, anywhere for anybody", function() {
      configuration = undefined;
      accessController = new AccessController(configuration, winston);

      Assert(accessController.isAccessAllowed("home.example.com", "/", "user1", ["group1", "group2"], WhitelistValue.NOT_WHITELISTED, false));
      Assert(accessController.isAccessAllowed("home.example.com", "/abc", "user1", ["group1", "group2"], WhitelistValue.NOT_WHITELISTED, false));
      Assert(accessController.isAccessAllowed("home.example.com", "/", "user2", ["group1", "group2"], WhitelistValue.NOT_WHITELISTED, false));
      Assert(accessController.isAccessAllowed("admin.example.com", "/", "user3", ["group3"], WhitelistValue.NOT_WHITELISTED, false));
    });
  });

  describe("configuration is not null", function () {
    beforeEach(function () {
      configuration = {
        default_policy: "deny",
        any: [],
        users: {},
        groups: {}
      };
      accessController = new AccessController(configuration, winston);
    });

    describe("check access control with default policy to deny", function () {
      beforeEach(function () {
        configuration.default_policy = "deny";
      });

      it("should deny access when no rule is provided", function () {
        Assert(!accessController.isAccessAllowed("home.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should control access when multiple domain matcher is provided", function () {
        configuration.users["user1"] = [{
          domain: "*.mail.example.com",
          policy: "allow",
          resources: [".*"]
        }];
        Assert(!accessController.isAccessAllowed("home.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("mx1.mail.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("mx1.server.mail.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("mail.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should allow access to all resources when resources is not provided", function () {
        configuration.users["user1"] = [{
          domain: "*.mail.example.com",
          policy: "allow"
        }];
        Assert(!accessController.isAccessAllowed("home.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("mx1.mail.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("mx1.server.mail.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("mail.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
      });

      describe("check user rules", function () {
        it("should allow access when user has a matching allowing rule", function () {
          configuration.users["user1"] = [{
            domain: "home.example.com",
            policy: "allow",
            resources: [".*"]
          }];
          Assert(accessController.isAccessAllowed("home.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(accessController.isAccessAllowed("home.example.com", "/another/resource", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("another.home.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        });

        it("should deny to other users", function () {
          configuration.users["user1"] = [{
            domain: "home.example.com",
            policy: "allow",
            resources: [".*"]
          }];
          Assert(!accessController.isAccessAllowed("home.example.com", "/", "user2", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("home.example.com", "/another/resource", "user2", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("another.home.example.com", "/", "user2", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        });

        it("should allow user access only to specific resources", function () {
          configuration.users["user1"] = [{
            domain: "home.example.com",
            policy: "allow",
            resources: ["/private/.*", "^/begin", "/end$"]
          }];
          Assert(!accessController.isAccessAllowed("home.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("home.example.com", "/private", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(accessController.isAccessAllowed("home.example.com", "/private/class", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(accessController.isAccessAllowed("home.example.com", "/middle/private/class", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));

          Assert(accessController.isAccessAllowed("home.example.com", "/begin", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("home.example.com", "/not/begin", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));

          Assert(accessController.isAccessAllowed("home.example.com", "/abc/end", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("home.example.com", "/abc/end/x", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        });

        it("should allow access to multiple domains", function () {
          configuration.users["user1"] = [{
            domain: "home.example.com",
            policy: "allow",
            resources: [".*"]
          }, {
            domain: "home1.example.com",
            policy: "allow",
            resources: [".*"]
          }, {
            domain: "home2.example.com",
            policy: "deny",
            resources: [".*"]
          }];
          Assert(accessController.isAccessAllowed("home.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(accessController.isAccessAllowed("home1.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("home2.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("home3.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        });

        it("should always apply latest rule", function () {
          configuration.users["user1"] = [{
            domain: "home.example.com",
            policy: "allow",
            resources: ["^/my/.*"]
          }, {
            domain: "home.example.com",
            policy: "deny",
            resources: ["^/my/private/.*"]
          }, {
            domain: "home.example.com",
            policy: "allow",
            resources: ["/my/private/resource"]
          }];

          Assert(accessController.isAccessAllowed("home.example.com", "/my/poney", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("home.example.com", "/my/private/duck", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(accessController.isAccessAllowed("home.example.com", "/my/private/resource", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        });
      });

      describe("check group rules", function () {
        it("should allow access when user is in group having a matching allowing rule", function () {
          configuration.groups["group1"] = [{
            domain: "home.example.com",
            policy: "allow",
            resources: ["^/$"]
          }];
          configuration.groups["group2"] = [{
            domain: "home.example.com",
            policy: "allow",
            resources: ["^/test$"]
          }, {
            domain: "home.example.com",
            policy: "deny",
            resources: ["^/private$"]
          }];
          Assert(accessController.isAccessAllowed("home.example.com", "/", "user1", ["group1", "group2", "group3"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(accessController.isAccessAllowed("home.example.com", "/test", "user1", ["group1", "group2", "group3"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("home.example.com", "/private", "user1", ["group1", "group2", "group3"], WhitelistValue.NOT_WHITELISTED, false));
          Assert(!accessController.isAccessAllowed("another.home.example.com", "/", "user1", ["group1", "group2", "group3"], WhitelistValue.NOT_WHITELISTED, false));
        });
      });
    });

    describe("check any rules", function () {
      it("should control access when any rules are defined", function () {
        configuration.any = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/public$"]
        }, {
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/private$"]
        }];
        Assert(accessController.isAccessAllowed("home.example.com", "/public", "user1", ["group1", "group2", "group3"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/private", "user1", ["group1", "group2", "group3"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/public", "user4", ["group5"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/private", "user4", ["group5"], WhitelistValue.NOT_WHITELISTED, false));
      });
    });

    describe("check access control with default policy to allow", function () {
      beforeEach(function () {
        configuration.default_policy = "allow";
      });

      it("should allow access to anything when no rule is provided", function () {
        Assert(accessController.isAccessAllowed("home.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/test", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/dev", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should deny access to one resource when defined", function () {
        configuration.users["user1"] = [{
          domain: "home.example.com",
          policy: "deny",
          resources: ["/test"]
        }];
        Assert(accessController.isAccessAllowed("home.example.com", "/", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/test", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/dev", "user1", ["group1"], WhitelistValue.NOT_WHITELISTED, false));
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
        configuration.any = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/public$", "^/$"]
        }];
        configuration.groups["dev"] = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/dev/?.*$"]
        }];
        configuration.groups["admins"] = [{
          domain: "home.example.com",
          policy: "allow",
          resources: [".*"]
        }];
        configuration.groups["admin-private"] = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/private/?.*"]
        }];
        configuration.users["john"] = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/private/john$"]
        }];
        configuration.users["harry"] = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/private/harry"]
        }, {
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/dev/b.*$"]
        }];

        Assert(accessController.isAccessAllowed("home.example.com", "/", "admin", ["admins"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/public", "admin", ["admins"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/dev", "admin", ["admins"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/dev/bob", "admin", ["admins"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/admin", "admin", ["admins"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/private/josh", "admin", ["admins"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/private/john", "admin", ["admins"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/private/harry", "admin", ["admins"], WhitelistValue.NOT_WHITELISTED, false));

        Assert(accessController.isAccessAllowed("home.example.com", "/", "john", ["dev", "admin-private"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/public", "john", ["dev", "admin-private"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/dev", "john", ["dev", "admin-private"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/dev/bob", "john", ["dev", "admin-private"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/admin", "john", ["dev", "admin-private"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/private/josh", "john", ["dev", "admin-private"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/private/john", "john", ["dev", "admin-private"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/private/harry", "john", ["dev", "admin-private"], WhitelistValue.NOT_WHITELISTED, false));

        Assert(accessController.isAccessAllowed("home.example.com", "/", "harry", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/public", "harry", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/dev", "harry", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/dev/bob", "harry", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/admin", "harry", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/private/josh", "harry", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/private/john", "harry", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/private/harry", "harry", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should control access when allowed at group level and denied at user level", function () {
        configuration.groups["dev"] = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/dev/?.*$"]
        }];
        configuration.users["john"] = [{
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/dev/bob$"]
        }];

        Assert(accessController.isAccessAllowed("home.example.com", "/dev/john", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/dev/bob", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should control access when allowed at 'any' level and denied at user level", function () {
        configuration.any = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/dev/?.*$"]
        }];
        configuration.users["john"] = [{
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/dev/bob$"]
        }];

        Assert(accessController.isAccessAllowed("home.example.com", "/dev/john", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/dev/bob", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should control access when allowed at 'any' level and denied at group level", function () {
        configuration.any = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/dev/?.*$"]
        }];
        configuration.groups["dev"] = [{
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/dev/bob$"]
        }];

        Assert(accessController.isAccessAllowed("home.example.com", "/dev/john", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/dev/bob", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should respect rules precedence", function () {
        // the priority from least to most is 'default_policy', 'all', 'group', 'user'
        // and the first rules in each category as a lower priority than the latest.
        // You can think of it that way: they override themselves inside each category.
        configuration.any = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/dev/?.*$"]
        }];
        configuration.groups["dev"] = [{
          domain: "home.example.com",
          policy: "deny",
          resources: ["^/dev/bob$"]
        }];
        configuration.users["john"] = [{
          domain: "home.example.com",
          policy: "allow",
          resources: ["^/dev/?.*$"]
        }];

        Assert(accessController.isAccessAllowed("home.example.com", "/dev/john", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/dev/bob", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
      });
    });

    describe("check whitelist access control with complete use case", function () {
      beforeEach(function () {
        configuration.default_policy = "deny";
      });

      it("should control whitelist access when allowed at group level and denied at user level", function () {
        configuration.groups["dev"] = [{
          domain: "home.example.com",
          policy: "allow",
          whitelist_policy: "allow",
          resources: ["^/dev/?.*$"]
        }];
        configuration.users["john"] = [{
          domain: "home.example.com",
          policy: "deny",
          whitelist_policy: "deny",
          resources: ["^/dev/bob$"]
        }];

        Assert(accessController.isAccessAllowed("home.example.com", "/dev/john", "john", ["dev"], WhitelistValue.WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/dev/bob", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should control whitelist access when allowed at 'any' level and denied at user level", function () {
        configuration.any = [{
          domain: "home.example.com",
          policy: "allow",
          whitelist_policy: "allow",
          resources: ["^/dev/?.*$"]
        }];
        configuration.users["john"] = [{
          domain: "home.example.com",
          policy: "deny",
          whitelist_policy: "deny",
          resources: ["^/dev/bob$"]
        }];

        Assert(accessController.isAccessAllowed("home.example.com", "/dev/john", "john", ["dev"], WhitelistValue.WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/dev/bob", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should control whitelist access when allowed at 'any' level and denied at group level", function () {
        configuration.any = [{
          domain: "home.example.com",
          policy: "allow",
          whitelist_policy: "allow",
          resources: ["^/dev/?.*$"]
        }];
        configuration.groups["dev"] = [{
          domain: "home.example.com",
          policy: "deny",
          whitelist_policy: "deny",
          resources: ["^/dev/bob$"]
        }];

        Assert(accessController.isAccessAllowed("home.example.com", "/dev/john", "john", ["dev"], WhitelistValue.WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/dev/bob", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should control access when user is whitelisted", function () {
        configuration.users["john"] = [{
          domain: "home.example.com",
          policy: "allow",
          whitelist_policy: "allow"
        }];

        Assert(accessController.isAccessAllowed("home.example.com", "/", "john", ["dev"], WhitelistValue.WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/", "john", ["dev"], WhitelistValue.WHITELISTED_AND_AUTHENTICATED_FIRSTFACTOR, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/", "john", ["dev"], WhitelistValue.WHITELISTED_AND_AUTHENTICATED_SECONDFACTOR, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/", "john", ["dev"], WhitelistValue.WHITELISTED_AND_AUTHENTICATED_SECONDFACTOR, true));
        Assert(!accessController.isAccessAllowed("home1.example.com", "/", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
        Assert(!accessController.isAccessAllowed("home2.example.com", "/", "john", ["dev"], WhitelistValue.NOT_WHITELISTED, false));
      });

      it("should control access when user is whitelisted and 'whitelist_policy' is denied", function () {
        configuration.users["john"] = [{
          domain: "home.example.com",
          policy: "allow",
          whitelist_policy: "deny"
        }];

        Assert(!accessController.isAccessAllowed("home.example.com", "/", "john", ["dev"], WhitelistValue.WHITELISTED, false));
        Assert(accessController.isAccessAllowed("home.example.com", "/", "john", ["dev"], WhitelistValue.WHITELISTED_AND_AUTHENTICATED_FIRSTFACTOR, false));
        Assert(!accessController.isAccessAllowed("home.example.com", "/", "john", ["dev"], WhitelistValue.WHITELISTED_AND_AUTHENTICATED_FIRSTFACTOR, true));
        Assert(accessController.isAccessAllowed("home.example.com", "/", "john", ["dev"], WhitelistValue.WHITELISTED_AND_AUTHENTICATED_SECONDFACTOR, true));
      });
    });
  });
});