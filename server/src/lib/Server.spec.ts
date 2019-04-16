
import Assert = require("assert");
import Sinon = require("sinon");
import nedb = require("nedb");
import winston = require("winston");
import speakeasy = require("speakeasy");
import u2f = require("u2f");
import session = require("express-session");
import { Configuration } from "./configuration/schema/Configuration";
import { GlobalDependencies } from "../../types/Dependencies";
import Server from "./Server";
import { LdapjsMock } from "./stubs/ldapjs.spec";


describe("Server", function () {
  let deps: GlobalDependencies;
  let sessionMock: Sinon.SinonSpy;
  let ldapjsMock: LdapjsMock;

  before(function () {
    sessionMock = Sinon.spy(session);
    ldapjsMock = new LdapjsMock();

    deps = {
      speakeasy: speakeasy,
      u2f: u2f,
      nedb: nedb,
      winston: winston,
      ldapjs: ldapjsMock as any,
      session: sessionMock as any,
      ConnectRedis: Sinon.spy(),
      Redis: Sinon.spy() as any
    };
  });


  it("should set cookie scope to domain set in the config", function () {
    const config: Configuration = {
      port: 8081,
      session: {
        domain: "example.com",
        secret: "secret"
      },
      authentication_backend: {
        ldap: {
          url: "http://ldap",
          user: "user",
          password: "password",
          base_dn: "dc=example,dc=com"
        },
      },
      notifier: {
        email: {
          username: "user@example.com",
          password: "password",
          sender: "test@authelia.com",
          service: "gmail"
        }
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

    const server = new Server(deps);
    server.start(config, deps)
      .then(function () {
        Assert(sessionMock.calledOnce);
        Assert.equal(sessionMock.getCall(0).args[0].cookie.domain, "example.com");
        server.stop();
      });
  });
});
