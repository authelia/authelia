import { ACLConfiguration, complete } from "./AclConfiguration";
import Assert = require("assert");

describe("configuration/schema/AclConfiguration", function() {
  it("should complete ACLConfiguration", function() {
    const configuration: ACLConfiguration = {};
    const [newConfiguration, errors] = complete(configuration);

    Assert.deepEqual(newConfiguration.default_policy, "bypass");
    Assert.deepEqual(newConfiguration.rules, []);
  });

  it("should return errors when subject is not good", function() {
    const configuration: ACLConfiguration = {
      default_policy: "deny",
      rules: [{
        domain: "dev.example.com",
        subject: "user:abc",
        policy: "bypass"
      }, {
        domain: "dev.example.com",
        subject: "user:def",
        policy: "bypass"
      }, {
        domain: "dev.example.com",
        subject: "badkey:abc",
        policy: "bypass"
      }]
    };
    const [newConfiguration, errors] = complete(configuration);

    Assert.deepEqual(errors, ["Rule 2 has wrong subject. It should be starting with user: or group:."]);
  });
});