import * as Assert from "assert";
import { UserConfiguration } from "../../src/types/Configuration";
import ConfigurationAdapter from "../../src/server/lib/ConfigurationAdapter";

describe("test config adapter", function() {
  function build_yaml_config(): UserConfiguration {
    const yaml_config = {
      port: 8080,
      ldap: {
        url: "http://ldap",
        base_dn: "cn=test,dc=example,dc=com",
        user: "user",
        password: "pass"
      },
      session: {
        domain: "example.com",
        secret: "secret",
        max_age: 40000
      },
      store_directory: "/mydirectory",
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

  it("should read the port from the yaml file", function() {
    const yaml_config = build_yaml_config();
    yaml_config.port = 7070;
    const config = ConfigurationAdapter.adapt(yaml_config);
    Assert.equal(config.port, 7070);
  });

  it("should default the port to 8080 if not provided", function() {
    const yaml_config = build_yaml_config();
    delete yaml_config.port;
    const config = ConfigurationAdapter.adapt(yaml_config);
    Assert.equal(config.port, 8080);
  });

  it("should get the ldap attributes", function() {
    const yaml_config = build_yaml_config();
    yaml_config.ldap = {
      url: "http://ldap",
      base_dn: "cn=test,dc=example,dc=com",
      additional_user_dn: "ou=users",
      user_name_attribute: "uid",
      user: "admin",
      password: "pass"
    };

    const config = ConfigurationAdapter.adapt(yaml_config);

    Assert.equal(config.ldap.url, "http://ldap");
    Assert.equal(config.ldap.additional_user_dn, "ou=users");
    Assert.equal(config.ldap.user_name_attribute, "uid");
    Assert.equal(config.ldap.user, "admin");
    Assert.equal(config.ldap.password, "pass");
  });

  it("should get the session attributes", function() {
    const yaml_config = build_yaml_config();
    yaml_config.session = {
      domain: "example.com",
      secret: "secret",
      expiration: 3600
    };
    const config = ConfigurationAdapter.adapt(yaml_config);
    Assert.equal(config.session.domain, "example.com");
    Assert.equal(config.session.secret, "secret");
    Assert.equal(config.session.expiration, 3600);
  });

  it("should get the log level", function() {
    const yaml_config = build_yaml_config();
    yaml_config.logs_level = "debug";
    const config = ConfigurationAdapter.adapt(yaml_config);
    Assert.equal(config.logs_level, "debug");
  });

  it("should get the notifier config", function() {
    const yaml_config = build_yaml_config();
    yaml_config.notifier = {
      gmail: {
        username: "user",
        password: "pass"
      }
    };
    const config = ConfigurationAdapter.adapt(yaml_config);
    Assert.deepEqual(config.notifier, {
      gmail: {
        username: "user",
        password: "pass"
      }
    });
  });

  it("should get the access_control config", function() {
    const yaml_config = build_yaml_config();
    yaml_config.access_control = {
      default: [],
      users: {},
      groups: {}
    };
    const config = ConfigurationAdapter.adapt(yaml_config);
    Assert.deepEqual(config.access_control, {
      default: [],
      users: {},
      groups: {}
    });
  });
});
