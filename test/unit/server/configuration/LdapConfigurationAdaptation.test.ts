import * as Assert from "assert";
import { UserConfiguration, LdapConfiguration } from "../../../../src/server/lib/configuration/Configuration";
import { ConfigurationAdapter } from "../../../../src/server/lib/configuration/ConfigurationAdapter";

describe("test ldap configuration adaptation", function () {
  function build_yaml_config(): UserConfiguration {
    const yaml_config = {
      port: 8080,
      ldap: {
        url: "http://ldap",
        base_dn: "dc=example,dc=com",
        additional_users_dn: "ou=users",
        additional_groups_dn: "ou=groups",
        user: "user",
        password: "pass"
      },
      session: {
        domain: "example.com",
        secret: "secret",
        max_age: 40000
      },
      storage: {
        local: {
          path: "/mydirectory"
        }
      },
      logs_level: "debug",
      notifier: {
        gmail: {
          username: "user",
          password: "password"
        }
      }
    };
    return yaml_config;
  }

  it("should adapt correctly while user only specify mandatory fields", function () {
    const yaml_config = build_yaml_config();
    yaml_config.ldap = {
      url: "http://ldap",
      base_dn: "dc=example,dc=com",
      user: "admin",
      password: "password"
    };

    const config = ConfigurationAdapter.adapt(yaml_config);
    const expectedConfig: LdapConfiguration = {
      url: "http://ldap",
      users_dn: "dc=example,dc=com",
      users_filter: "cn={0}",
      groups_dn: "dc=example,dc=com",
      groups_filter: "member={0}",
      group_name_attribute: "cn",
      mail_attribute: "mail",
      user: "admin",
      password: "password"
    };

    Assert.deepEqual(config.ldap, expectedConfig);
  });

  it("should adapt correctly while user specify every fields", function () {
    const yaml_config = build_yaml_config();
    yaml_config.ldap = {
      url: "http://ldap-server",
      base_dn: "dc=example,dc=com",
      additional_users_dn: "ou=users",
      users_filter: "uid={0}",
      additional_groups_dn: "ou=groups",
      groups_filter: "uniqueMember={0}",
      mail_attribute: "email",
      group_name_attribute: "groupName",
      user: "admin2",
      password: "password2"
    };

    const config = ConfigurationAdapter.adapt(yaml_config);
    const expectedConfig: LdapConfiguration = {
      url: "http://ldap-server",
      users_dn: "ou=users,dc=example,dc=com",
      users_filter: "uid={0}",
      groups_dn: "ou=groups,dc=example,dc=com",
      groups_filter: "uniqueMember={0}",
      mail_attribute: "email",
      group_name_attribute: "groupName",
      user: "admin2",
      password: "password2"
    };

    Assert.deepEqual(config.ldap, expectedConfig);
  });
});
