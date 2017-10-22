import { Validator } from "../../src/lib/configuration/Validator";
import Assert = require("assert");

describe("test validator", function () {
  it("should validate wrong user configurations", function () {
    // Some examples
    Assert.deepStrictEqual(Validator.isValid({}), [
      "data should have required property 'ldap'",
      "data should have required property 'notifier'",
      "data should have required property 'regulation'",
      "data should have required property 'session'",
      "data should have required property 'storage'"
    ]);

    Assert.deepStrictEqual(Validator.isValid({
      ldap: {},
      notifier: {},
      regulation: {},
      session: {},
      storage: {}
    }), [
        "data.ldap should have required property 'base_dn'",
        "data.ldap should have required property 'password'",
        "data.ldap should have required property 'url'",
        "data.ldap should have required property 'user'",
        "data.regulation should have required property 'ban_time'",
        "data.regulation should have required property 'find_time'",
        "data.regulation should have required property 'max_retries'",
        "data.session should have required property 'secret'",
        "Storage must be either 'local' or 'mongo'",
        "A notifier needs to be declared when server is used with two-factor"
      ]);

    Assert.deepStrictEqual(Validator.isValid({
      ldap: {
        base_dn: "dc=example,dc=com",
        password: "password",
        url: "ldap://ldap",
        user: "user"
      },
      notifier: {
        abcd: []
      },
      regulation: {
        ban_time: 120,
        find_time: 30,
        max_retries: 3
      },
      session: {
        secret: "unsecure_secret"
      },
      storage: {
        abc: {}
      }
    }), [
        "data.storage has unknown key 'abc'",
        "Notifier must be either 'filesystem', 'gmail' or 'smtp'"
      ]);
  });

  it("should validate correct user configurations", function () {
    Assert.deepStrictEqual(Validator.isValid({
      ldap: {
        base_dn: "dc=example,dc=com",
        password: "password",
        url: "ldap://ldap",
        user: "user"
      },
      notifier: {
        gmail: {
          username: "user@gmail.com",
          password: "pass",
          sender: "admin@example.com"
        }
      },
      regulation: {
        ban_time: 120,
        find_time: 30,
        max_retries: 3
      },
      session: {
        secret: "unsecure_secret"
      },
      storage: {
        local: {
          path: "/var/lib/authelia"
        }
      }
    }), []);
  });

  it("should return false when notifier is not defined while there is at least \
one second factor enabled sub-domain", function () {
      const options1 = {
        ldap: {
          base_dn: "dc=example,dc=com",
          password: "password",
          url: "ldap://ldap",
          user: "user"
        },
        authentication_methods: {
          default_method: "two_factor"
        },
        notifier: {},
        regulation: {
          ban_time: 120,
          find_time: 30,
          max_retries: 3
        },
        session: {
          secret: "unsecure_secret"
        },
        storage: {
          local: {
            path: "/var/lib/authelia"
          }
        }
      };
      const options2 = {
        ldap: {
          base_dn: "dc=example,dc=com",
          password: "password",
          url: "ldap://ldap",
          user: "user"
        },
        authentication_methods: {
          default_method: "two_factor"
        },
        notifier: {
          gmail: {
            username: "user@gmail.com",
            password: "pass",
            sender: "admin@example.com"
          }
        },
        regulation: {
          ban_time: 120,
          find_time: 30,
          max_retries: 3
        },
        session: {
          secret: "unsecure_secret"
        },
        storage: {
          local: {
            path: "/var/lib/authelia"
          }
        }
      };
      const options3 = {
        ldap: {
          base_dn: "dc=example,dc=com",
          password: "password",
          url: "ldap://ldap",
          user: "user"
        },
        authentication_methods: {
          default_method: "single_factor"
        },
        notifier: {},
        regulation: {
          ban_time: 120,
          find_time: 30,
          max_retries: 3
        },
        session: {
          secret: "unsecure_secret"
        },
        storage: {
          local: {
            path: "/var/lib/authelia"
          }
        }
      };
      Assert.deepStrictEqual(Validator.isValid(options1), ["A notifier needs to be declared when server is used with two-factor"]);
      Assert.deepStrictEqual(Validator.isValid(options2), []);
      Assert.deepStrictEqual(Validator.isValid(options3), []);
    });
});