import Assert = require("assert");
import { RegulationConfiguration, complete } from "./RegulationConfiguration";

describe("configuration/schema/RegulationConfiguration", function() {
  it("should return default regulation configuration", function() {
    const configuration: RegulationConfiguration = {};
    const newConfiguration = complete(configuration);

    Assert.equal(newConfiguration.ban_time, 300);
    Assert.equal(newConfiguration.find_time, 120);
    Assert.equal(newConfiguration.max_retries, 3);
  });
});