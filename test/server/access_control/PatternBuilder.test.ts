
import assert = require("assert");
import winston = require("winston");

import PatternBuilder from "../../../src/server/lib/access_control/PatternBuilder";
import { ACLConfiguration } from "../../../src/types/Configuration";

describe("test access control manager", function () {
  describe("test access control pattern builder when no configuration is provided", () => {
    it("should allow access to the user", () => {
      const patternBuilder = new PatternBuilder(undefined, winston);

      const allowed_domains = patternBuilder.getAllowedDomains("user", ["group1"]);
      assert.deepEqual(allowed_domains, ["*"]);
    });
  });

  describe("test access control pattern builder", function () {
    let patternBuilder: PatternBuilder;
    let configuration: ACLConfiguration;


    beforeEach(() => {
      configuration = {
        default: [],
        users: {},
        groups: {}
      };
      patternBuilder = new PatternBuilder(configuration, winston);
    });

    it("should deny all if nothing is defined in the config", function () {
      const allowed_domains = patternBuilder.getAllowedDomains("user", ["group1", "group2"]);
      assert.deepEqual(allowed_domains, []);
    });

    it("should allow domain test.example.com to all users if defined in" +
      " default policy", function () {
        configuration.default = ["test.example.com"];
        const allowed_domains = patternBuilder.getAllowedDomains("user", ["group1", "group2"]);
        assert.deepEqual(allowed_domains, ["test.example.com"]);
      });

    it("should allow domain test.example.com to all users in group mygroup", function () {
      const allowed_domains0 = patternBuilder.getAllowedDomains("user", ["group1", "group1"]);
      assert.deepEqual(allowed_domains0, []);

      configuration.groups = {
        mygroup: ["test.example.com"]
      };

      const allowed_domains1 = patternBuilder.getAllowedDomains("user", ["group1", "group2"]);
      assert.deepEqual(allowed_domains1, []);

      const allowed_domains2 = patternBuilder.getAllowedDomains("user", ["group1", "mygroup"]);
      assert.deepEqual(allowed_domains2, ["test.example.com"]);
    });

    it("should allow domain test.example.com based on per user config", function () {
      const allowed_domains0 = patternBuilder.getAllowedDomains("user", ["group1"]);
      assert.deepEqual(allowed_domains0, []);

      configuration.users = {
        user1: ["test.example.com"]
      };

      const allowed_domains1 = patternBuilder.getAllowedDomains("user", ["group1", "mygroup"]);
      assert.deepEqual(allowed_domains1, []);

      const allowed_domains2 = patternBuilder.getAllowedDomains("user1", ["group1", "mygroup"]);
      assert.deepEqual(allowed_domains2, ["test.example.com"]);
    });

    it("should allow domains from user and groups", function () {
      configuration.groups = {
        group2: ["secret.example.com", "secret1.example.com"]
      };
      configuration.users = {
        user: ["test.example.com"]
      };

      const allowed_domains0 = patternBuilder.getAllowedDomains("user", ["group1", "group2"]);
      assert.deepEqual(allowed_domains0, [
        "secret.example.com",
        "secret1.example.com",
        "test.example.com",
      ]);
    });

    it("should allow domains from several groups", function () {
      configuration.groups = {
        group1: ["secret2.example.com"],
        group2: ["secret.example.com", "secret1.example.com"]
      };

      const allowed_domains0 = patternBuilder.getAllowedDomains("user", ["group1", "group2"]);
      assert.deepEqual(allowed_domains0, [
        "secret2.example.com",
        "secret.example.com",
        "secret1.example.com",
      ]);
    });

    it("should allow domains from several groups and default policy", function () {
      configuration.default = ["home.example.com"];
      configuration.groups = {
        group1: ["secret2.example.com"],
        group2: ["secret.example.com", "secret1.example.com"]
      };

      const allowed_domains0 = patternBuilder.getAllowedDomains("user", ["group1", "group2"]);
      assert.deepEqual(allowed_domains0, [
        "home.example.com",
        "secret2.example.com",
        "secret.example.com",
        "secret1.example.com",
      ]);
    });
  });
});
