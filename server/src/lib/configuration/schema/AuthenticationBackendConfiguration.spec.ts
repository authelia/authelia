import { AuthenticationBackendConfiguration, complete } from "./AuthenticationBackendConfiguration";
import Assert = require("assert");

describe("configuration/schema/AuthenticationBackendConfiguration", function() {
  it("should ensure there is at least one key", function() {
    const configuration: AuthenticationBackendConfiguration = {} as any;
    const [newConfiguration, error] = complete(configuration);

    Assert.equal(error, "Authentication backend must have one of the following keys:`ldap` or `file`");
  });
});