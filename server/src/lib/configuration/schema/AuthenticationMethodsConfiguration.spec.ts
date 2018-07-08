import Assert = require("assert");
import { AuthenticationMethodsConfiguration, complete } from "./AuthenticationMethodsConfiguration";

describe("configuration/schema/AuthenticationMethodsConfiguration", function() {
  it("should ensure at least one key is provided", function() {
    const configuration: AuthenticationMethodsConfiguration = {};
    const newConfiguration = complete(configuration);

    Assert.deepEqual(newConfiguration.default_method, "two_factor");
    Assert.deepEqual(newConfiguration.per_subdomain_methods, []);
  });
});