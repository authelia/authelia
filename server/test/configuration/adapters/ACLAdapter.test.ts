import { ACLAdapter } from "../../../src/lib/configuration/adapters/ACLAdapter";
import Assert = require("assert");

describe("test ACL configuration adapter", function () {

  describe("bad default_policy", function () {
    it("should adapt a configuration missing default_policy", function () {
      const userConfiguration: any = {
        any: [],
        groups: {},
        users: {}
      };

      const appConfiguration = ACLAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_policy: "deny",
        any: [],
        groups: {},
        users: {}
      });
    });

    it("should adapt a configuration with bad default_policy value", function () {
      const userConfiguration: any = {
        default_policy: "anything", // it should be 'allow' or 'deny'
        any: [],
        groups: {},
        users: {}
      };

      const appConfiguration = ACLAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_policy: "deny",
        any: [],
        groups: {},
        users: {}
      });
    });

    it("should adapt a configuration with bad default_policy type", function () {
      const userConfiguration: any = {
        default_policy: {}, // it should be 'allow' or 'deny'
        any: [],
        groups: {},
        users: {}
      };

      const appConfiguration = ACLAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_policy: "deny",
        any: [],
        groups: {},
        users: {}
      });
    });
  });

  describe("bad any", function () {
    it("should adapt a configuration missing any key", function () {
      const userConfiguration: any = {
        default_policy: "deny",
        groups: {},
        users: {}
      };

      const appConfiguration = ACLAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_policy: "deny",
        any: [],
        groups: {},
        users: {}
      });
    });

    it("should adapt a configuration with any not being an array", function () {
      const userConfiguration: any = {
        default_policy: "deny",
        any: "abc",
        groups: {},
        users: {}
      };

      const appConfiguration = ACLAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_policy: "deny",
        any: [],
        groups: {},
        users: {}
      });
    });
  });

  describe("bad groups", function () {
    it("should adapt a configuration missing groups key", function () {
      const userConfiguration: any = {
        default_policy: "deny",
        any: [],
        users: {}
      };

      const appConfiguration = ACLAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_policy: "deny",
        any: [],
        groups: {},
        users: {}
      });
    });

    it("should adapt configuration with groups being of wrong type", function () {
      const userConfiguration: any = {
        default_policy: "deny",
        any: [],
        groups: [],
        users: {}
      };

      const appConfiguration = ACLAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_policy: "deny",
        any: [],
        groups: {},
        users: {}
      });
    });
  });

  describe("bad users", function () {
    it("should adapt a configuration missing users key", function () {
      const userConfiguration: any = {
        default_policy: "deny",
        any: [],
        groups: {}
      };

      const appConfiguration = ACLAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_policy: "deny",
        any: [],
        groups: {},
        users: {}
      });
    });

    it("should adapt a configuration with users being of wrong type", function () {
      const userConfiguration: any = {
        default_policy: "deny",
        any: [],
        groups: {},
        users: []
      };

      const appConfiguration = ACLAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_policy: "deny",
        any: [],
        groups: {},
        users: {}
      });
    });
  });
});