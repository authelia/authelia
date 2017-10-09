import { Validator } from "../../src/lib/configuration/Validator";
import Assert = require("assert");

describe.only("test validator", function() {
  it("should validate a correct user configuration", function() {
    Assert(Validator.validate({
      ldap: {}
    }));
  });
});