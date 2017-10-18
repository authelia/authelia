import { AuthenticationMethodsAdapter } from "../../../src/lib/configuration/adapters/AuthenticationMethodsAdapter";
import Assert = require("assert");

describe("test authentication methods configuration adapter", function () {
  describe("no authentication methods defined", function () {
    it("should adapt a configuration when no authentication methods config is defined", function () {
      const userConfiguration: any = undefined;

      const appConfiguration = AuthenticationMethodsAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_method: "two_factor",
        per_subdomain_methods: {}
      });
    });
  });

  describe("partial authentication methods config", function() {
    it("should adapt a configuration when default_method is not defined", function () {
      const userConfiguration: any = {
        per_subdomain_methods: {
          "example.com": "single_factor"
        }
      };

      const appConfiguration = AuthenticationMethodsAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_method: "two_factor",
        per_subdomain_methods: {
          "example.com": "single_factor"
        }
      });
    });

    it("should adapt a configuration when per_subdomain_methods is not defined", function () {
      const userConfiguration: any = {
        default_method: "single_factor"
      };

      const appConfiguration = AuthenticationMethodsAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_method: "single_factor",
        per_subdomain_methods: {}
      });
    });

    it("should adapt a configuration when per_subdomain_methods has wrong type", function () {
      const userConfiguration: any = {
        default_method: "single_factor",
        per_subdomain_methods: []
      };

      const appConfiguration = AuthenticationMethodsAdapter.adapt(userConfiguration);
      Assert.deepStrictEqual(appConfiguration, {
        default_method: "single_factor",
        per_subdomain_methods: {}
      });
    });
  });
});
