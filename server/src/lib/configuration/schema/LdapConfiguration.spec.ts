import Assert = require("assert");
import { LdapConfiguration, complete } from "./LdapConfiguration";

describe("configuration/schema/AuthenticationMethodsConfiguration", function() {
  it("should ensure at least one key is provided", function() {
    const configuration: LdapConfiguration = {
      url: "ldap.example.com",
      base_dn: "dc=example,dc=com",
      user: "admin",
      password: "password"
    };
    const newConfiguration = complete(configuration);

    Assert.deepEqual(newConfiguration, {
      url: "ldap.example.com",
      base_dn: "dc=example,dc=com",
      user: "admin",
      password: "password",
      users_filter: "cn={0}",
      group_name_attribute: "cn",
      groups_filter: "member={dn}",
      mail_attribute: "mail"
    });
  });
});