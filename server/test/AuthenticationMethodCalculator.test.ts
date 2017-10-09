import { AuthenticationMethodCalculator } from "../src/lib/AuthenticationMethodCalculator";
import { AuthenticationMethodsConfiguration } from "../src/lib/configuration/Configuration";
import Assert = require("assert");

describe("test authentication method calculator", function() {
  it("should return default method when sub domain not overriden", function() {
    const options1: AuthenticationMethodsConfiguration = {
      default_method: "two_factor",
      per_subdomain_methods: {}
    };
    const options2: AuthenticationMethodsConfiguration = {
      default_method: "basic_auth",
      per_subdomain_methods: {}
    };
    const calculator1 = new AuthenticationMethodCalculator(options1);
    const calculator2 = new AuthenticationMethodCalculator(options2);
    Assert.equal(calculator1.compute("www.example.com"), "two_factor");
    Assert.equal(calculator2.compute("www.example.com"), "basic_auth");
  });

  it("should return overridden method when sub domain method is defined", function() {
    const options1: AuthenticationMethodsConfiguration = {
      default_method: "two_factor",
      per_subdomain_methods: {
        "www.example.com": "basic_auth"
      }
    };
    const calculator1 = new AuthenticationMethodCalculator(options1);
    Assert.equal(calculator1.compute("www.example.com"), "basic_auth");
  });
});