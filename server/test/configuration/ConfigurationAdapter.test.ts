import * as Assert from "assert";
import {
  UserConfiguration,
  LdapConfiguration, ACLConfiguration
} from "../../src/lib/configuration/Configuration";
import { ConfigurationAdapter } from "../../src/lib/configuration/ConfigurationAdapter";

describe("test config adapter", function () {
  function build_yaml_config(): UserConfiguration {
    const yaml_config: UserConfiguration = {
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
        expiration: 40000
      },
      storage: {
        local: {
          path: "/mydirectory"
        }
      },
      regulation: {
        max_retries: 3,
        find_time: 5 * 60,
        ban_time: 5 * 60
      },
      logs_level: "debug",
      notifier: {
        gmail: {
          username: "user",
          password: "password",
          sender: "admin@example.com"
        }
      }
    };
    return yaml_config;
  }

  describe("port", function () {
    it("should read the port from the yaml file", function () {
      const yaml_config = build_yaml_config();
      yaml_config.port = 7070;
      const config = ConfigurationAdapter.adapt(yaml_config);
      Assert.equal(config.port, 7070);
    });

    it("should default the port to 8080 if not provided", function () {
      const yaml_config = build_yaml_config();
      delete yaml_config.port;
      const config = ConfigurationAdapter.adapt(yaml_config);
      Assert.equal(config.port, 8080);
    });
  });

  it("should get the session attributes", function () {
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

  it("should get the log level", function () {
    const yaml_config = build_yaml_config();
    yaml_config.logs_level = "debug";
    const config = ConfigurationAdapter.adapt(yaml_config);
    Assert.equal(config.logs_level, "debug");
  });

  it("should get the notifier config", function () {
    const yaml_config = build_yaml_config();
    yaml_config.notifier = {
      gmail: {
        username: "user",
        password: "pass",
        sender: "admin@example.com"
      }
    };
    const config = ConfigurationAdapter.adapt(yaml_config);
    Assert.deepEqual(config.notifier, {
      gmail: {
        username: "user",
        password: "pass",
        sender: "admin@example.com"
      }
    });
  });

  describe("access_control", function() {
    it("should adapt access_control when it is already ok", function () {
      const yaml_config = build_yaml_config();
      yaml_config.access_control = {
        default_policy: "deny",
        any: [{
          domain: "public.example.com",
          policy: "allow"
        }],
        users: {
          "user": [{
            domain: "www.example.com",
            policy: "allow"
          }]
        },
        groups: {}
      };
      const config = ConfigurationAdapter.adapt(yaml_config);
      Assert.deepEqual(config.access_control, {
        default_policy: "deny",
        any: [{
          domain: "public.example.com",
          policy: "allow"
        }],
        users: {
          "user": [{
            domain: "www.example.com",
            policy: "allow"
          }]
        },
        groups: {}
      } as ACLConfiguration);
    });


    it("should adapt access_control when it is empty", function () {
      const yaml_config = build_yaml_config();
      yaml_config.access_control = {} as any;
      const config = ConfigurationAdapter.adapt(yaml_config);
      Assert.deepEqual(config.access_control, {
        default_policy: "deny",
        any: [],
        users: {},
        groups: {}
      } as ACLConfiguration);
    });
  });
});
