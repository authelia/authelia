import { SessionConfigurationBuilder } from "./SessionConfigurationBuilder";
import { Configuration } from "./schema/Configuration";
import { GlobalDependencies } from "../../../types/Dependencies";

import ExpressSession = require("express-session");
import Sinon = require("sinon");
import Assert = require("assert");

describe("configuration/SessionConfigurationBuilder", function () {
  const configuration: Configuration = {
    access_control: {
      default_policy: "deny",
      rules: []
    },
    totp: {
      issuer: "authelia.com"
    },
    authentication_backend: {
      ldap: {
        url: "ldap://ldap",
        user: "user",
        base_dn: "dc=example,dc=com",
        password: "password",
        additional_groups_dn: "ou=groups",
        additional_users_dn: "ou=users",
        group_name_attribute: "",
        groups_filter: "",
        mail_attribute: "",
        users_filter: ""
      },
    },
    logs_level: "debug",
    notifier: {
      filesystem: {
        filename: "/test"
      }
    },
    port: 8080,
    session: {
      name: "authelia_session",
      domain: "example.com",
      expiration: 3600,
      secret: "secret"
    },
    regulation: {
      max_retries: 3,
      ban_time: 5 * 60,
      find_time: 5 * 60
    },
    storage: {
      local: {
        in_memory: true
      }
    }
  };

  const deps: GlobalDependencies = {
    ConnectRedis: Sinon.spy() as any,
    ldapjs: Sinon.spy() as any,
    nedb: Sinon.spy() as any,
    session: Sinon.spy() as any,
    speakeasy: Sinon.spy() as any,
    u2f: Sinon.spy() as any,
    winston: Sinon.spy() as any,
    Redis: Sinon.spy() as any
  };

  it("should return session options without redis options", function () {
    const options = SessionConfigurationBuilder.build(configuration, deps);
    const expectedOptions = {
      name: "authelia_session",
      secret: "secret",
      resave: false,
      saveUninitialized: true,
      cookie: {
        secure: true,
        httpOnly: true,
        maxAge: 3600,
        domain: "example.com"
      }
    };

    Assert.deepEqual(expectedOptions, options);
  });

  it("should return session options with redis options", function () {
    configuration.session["redis"] = {
      host: "redis.example.com",
      port: 6379
    };
    const RedisStoreMock = Sinon.spy();
    const redisClient = Sinon.mock().returns({ on: Sinon.spy() });

    deps.ConnectRedis = Sinon.stub().returns(RedisStoreMock) as any;
    deps.Redis = {
      createClient: Sinon.mock().returns(redisClient)
    } as any;

    const options = SessionConfigurationBuilder.build(configuration, deps);

    const expectedOptions: ExpressSession.SessionOptions = {
      secret: "secret",
      resave: false,
      saveUninitialized: true,
      name: "authelia_session",
      cookie: {
        secure: true,
        httpOnly: true,
        maxAge: 3600,
        domain: "example.com"
      },
      store: Sinon.match.object as any
    };

    Assert((deps.ConnectRedis as Sinon.SinonStub).calledWith(deps.session));
    Assert.equal(options.secret, expectedOptions.secret);
    Assert.equal(options.resave, expectedOptions.resave);
    Assert.equal(options.saveUninitialized, expectedOptions.saveUninitialized);
    Assert.deepEqual(options.cookie, expectedOptions.cookie);
    Assert(options.store != undefined);
  });

  it("should return session options with redis password", function () {
    configuration.session["redis"] = {
      host: "redis.example.com",
      port: 6379,
      password: "authelia_pass"
    };
    const RedisStoreMock = Sinon.spy();
    deps.ConnectRedis = Sinon.stub().returns(RedisStoreMock);

    SessionConfigurationBuilder.build(configuration, deps);

    Assert(RedisStoreMock.calledWith({
      host: "redis.example.com",
      port: 6379,
      pass: "authelia_pass",
      logErrors: true,
    }));
  });
});