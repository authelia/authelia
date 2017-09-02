
import assert = require("assert");
import Sinon = require("sinon");
import nedb = require("nedb");
import express = require("express");
import winston = require("winston");
import speakeasy = require("speakeasy");
import u2f = require("u2f");
import nodemailer = require("nodemailer");
import session = require("express-session");

import { AppConfiguration, UserConfiguration } from "../../../src/server/lib/configuration/Configuration";
import { GlobalDependencies, Nodemailer } from "../../../src/types/Dependencies";
import Server from "../../../src/server/lib/Server";


describe("test server configuration", function () {
  let deps: GlobalDependencies;
  let sessionMock: Sinon.SinonSpy;

  before(function () {
    const transporter = {
      sendMail: Sinon.stub().yields()
    };

    const createTransport = Sinon.stub(nodemailer, "createTransport");
    createTransport.returns(transporter);

    sessionMock = Sinon.spy(session);

    deps = {
      nodemailer: nodemailer,
      speakeasy: speakeasy,
      u2f: u2f,
      nedb: nedb,
      winston: winston,
      ldapjs: {
        createClient: Sinon.spy(function () {
          return {
            on: Sinon.spy(),
            bind: Sinon.spy(),
          };
        })
      },
      session: sessionMock as any,
      ConnectRedis: Sinon.spy(),
      dovehash: Sinon.spy() as any
    };
  });


  it("should set cookie scope to domain set in the config", function () {
    const config: UserConfiguration = {
      session: {
        domain: "example.com",
        secret: "secret"
      },
      ldap: {
        url: "http://ldap",
        user: "user",
        password: "password",
        base_dn: "dc=example,dc=com"
      },
      notifier: {
        gmail: {
          username: "user@example.com",
          password: "password"
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

    const server = new Server();
    server.start(config, deps);

    assert(sessionMock.calledOnce);
    assert.equal(sessionMock.getCall(0).args[0].cookie.domain, "example.com");
  });
});
