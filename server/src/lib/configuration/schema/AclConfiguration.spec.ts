import { ACLConfiguration, complete } from "./AclConfiguration";
import Assert = require("assert");

describe("configuration/schema/AclConfiguration", function() {
  it("should complete ACLConfiguration", function() {
    const configuration: ACLConfiguration = {};
    const newConfiguration = complete(configuration);

    Assert.deepEqual(newConfiguration.default_policy, "bypass");
    Assert.deepEqual(newConfiguration.any, []);
    Assert.deepEqual(newConfiguration.groups, {});
    Assert.deepEqual(newConfiguration.users, {});
  });
});