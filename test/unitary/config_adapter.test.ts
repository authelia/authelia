import * as mocha from "mocha";
import * as Assert from "assert";

const config_adapter = require("../../src/lib/config_adapter");

describe("test config adapter", function() {
  function build_yaml_config(): any {
    const yaml_config = {
      port: 8080,
      ldap: {},
      session: {
        domain: "example.com",
        secret: "secret",
        max_age: 40000
      },
      store_directory: "/mydirectory",
      logs_level: "debug"
    };
    return yaml_config;
  }

  it("should read the port from the yaml file", function() {
    const yaml_config = build_yaml_config();
    yaml_config.port = 7070;
    const config = config_adapter(yaml_config);
    Assert.equal(config.port, 7070);
  });

  it("should default the port to 8080 if not provided", function() {
    const yaml_config = build_yaml_config();
    delete yaml_config.port;
    const config = config_adapter(yaml_config);
    Assert.equal(config.port, 8080);
  });

  it("should get the ldap attributes", function() {
    const yaml_config = build_yaml_config();
    yaml_config.ldap = {
      url: "http://ldap",
      user_search_base: "ou=groups,dc=example,dc=com",
      user_search_filter: "uid",
      user: "admin",
      password: "pass"
    };

    const config = config_adapter(yaml_config);

    Assert.equal(config.ldap.url, "http://ldap");
    Assert.equal(config.ldap.user_search_base, "ou=groups,dc=example,dc=com");
    Assert.equal(config.ldap.user_search_filter, "uid");
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
    const config = config_adapter(yaml_config);
    Assert.equal(config.session_domain, "example.com");
    Assert.equal(config.session_secret, "secret");
    Assert.equal(config.session_max_age, 3600);
  });

  it("should get the log level", function() {
    const yaml_config = build_yaml_config();
    yaml_config.logs_level = "debug";
    const config = config_adapter(yaml_config);
    Assert.equal(config.logs_level, "debug");
  });

  it("should get the notifier config", function() {
    const yaml_config = build_yaml_config();
    yaml_config.notifier = "notifier";
    const config = config_adapter(yaml_config);
    Assert.equal(config.notifier, "notifier");
  });

  it("should get the access_control config", function() {
    const yaml_config = build_yaml_config();
    yaml_config.access_control = "access_control";
    const config = config_adapter(yaml_config);
    Assert.equal(config.access_control, "access_control");
  });
});
