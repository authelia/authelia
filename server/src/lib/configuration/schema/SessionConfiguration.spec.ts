import Assert = require("assert");
import { SessionConfiguration, complete } from "./SessionConfiguration";

describe("configuration/schema/SessionConfiguration", function() {
  it("should return default regulation configuration", function() {
    const configuration: SessionConfiguration = {
      domain: "example.com",
      secret: "unsecure_secret"
    };
    const newConfiguration = complete(configuration);

    Assert.equal(newConfiguration.name, 'authelia_session');
    Assert.equal(newConfiguration.expiration, 3600000);
    Assert.equal(newConfiguration.inactivity, undefined);
  });
});