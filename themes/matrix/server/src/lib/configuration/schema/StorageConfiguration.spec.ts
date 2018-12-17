import Assert = require("assert");
import { StorageConfiguration, complete } from "./StorageConfiguration";

describe("configuration/schema/StorageConfiguration", function() {
  it("should return default regulation configuration", function() {
    const configuration: StorageConfiguration = {};
    const newConfiguration = complete(configuration);

    Assert.deepEqual(newConfiguration, {
      local: {
        in_memory: true
      }
    });
  });
});