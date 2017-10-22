import { MethodCalculator }
  from "../../src/lib/authentication/MethodCalculator";
import { AuthenticationMethodsConfiguration }
  from "../../src/lib/configuration/Configuration";
import Assert = require("assert");

describe("test MethodCalculator", function () {
  describe("test compute method", function () {
    it("should return default method when sub domain not overriden",
      function () {
        const options1: AuthenticationMethodsConfiguration = {
          default_method: "two_factor",
          per_subdomain_methods: {}
        };
        const options2: AuthenticationMethodsConfiguration = {
          default_method: "single_factor",
          per_subdomain_methods: {}
        };
        Assert.equal(MethodCalculator.compute(options1, "www.example.com"),
          "two_factor");
        Assert.equal(MethodCalculator.compute(options2, "www.example.com"),
          "single_factor");
      });

    it("should return overridden method when sub domain method is defined",
      function () {
        const options1: AuthenticationMethodsConfiguration = {
          default_method: "two_factor",
          per_subdomain_methods: {
            "www.example.com": "single_factor"
          }
        };
        Assert.equal(MethodCalculator.compute(options1, "www.example.com"),
          "single_factor");
        Assert.equal(MethodCalculator.compute(options1, "home.example.com"),
          "two_factor");
      });
  });

  describe("test isSingleFactorOnlyMode method", function () {
    it("should return true when default domains and all domains are single_factor",
      function () {
        const options: AuthenticationMethodsConfiguration = {
          default_method: "single_factor",
          per_subdomain_methods: {
            "www.example.com": "single_factor"
          }
        };
        Assert(MethodCalculator.isSingleFactorOnlyMode(options));
      });

    it("should return false when default domains is single_factor and at least one sub-domain is two_factor", function () {
      const options: AuthenticationMethodsConfiguration = {
        default_method: "single_factor",
        per_subdomain_methods: {
          "www.example.com": "two_factor",
          "home.example.com": "single_factor"
        }
      };
      Assert(!MethodCalculator.isSingleFactorOnlyMode(options));
    });

    it("should return false when default domains is two_factor", function () {
      const options: AuthenticationMethodsConfiguration = {
        default_method: "two_factor",
        per_subdomain_methods: {
          "www.example.com": "single_factor",
          "home.example.com": "single_factor"
        }
      };
      Assert(!MethodCalculator.isSingleFactorOnlyMode(options));
    });
  });
});