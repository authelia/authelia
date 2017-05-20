
import Server from "../../src/lib/Server";

import { UserConfiguration } from "../../src/lib/Configuration";
import { GlobalDependencies } from "../../src/lib/Dependencies";
import * as express from "express";

const sinon = require("sinon");
const assert = require("assert");

describe("test server configuration", function () {
  let deps: GlobalDependencies;

  before(function () {
    const transporter = {
      sendMail: sinon.stub().yields()
    };

    const nodemailer = {
      createTransport: sinon.spy(function () {
        return transporter;
      })
    };

    deps = {
      nodemailer: nodemailer,
      speakeasy: sinon.spy(),
      u2f: sinon.spy(),
      nedb: require("nedb"),
      winston: sinon.spy(),
      ldapjs: {
        createClient: sinon.spy(function () {
          return { on: sinon.spy() };
        })
      },
      session: sinon.spy(function () {
        return function (req: express.Request, res: express.Response, next: express.NextFunction) { next(); };
      })
    };
  });


  it("should set cookie scope to domain set in the config", function () {
    const config = {
      notifier: {
        gmail: {
          username: "user@example.com",
          password: "password"
        }
      },
      session: {
        domain: "example.com",
        secret: "secret"
      },
      ldap: {
        url: "http://ldap",
        base_dn: "cn=test,dc=example,dc=com",
        user: "user",
        password: "password"
      }
    };

    const server = new Server();
    server.start(config, deps);

    assert(deps.session.calledOnce);
    assert.equal(deps.session.getCall(0).args[0].cookie.domain, "example.com");
  });
});
