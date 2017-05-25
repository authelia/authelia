
import assert = require("assert");
import sinon = require ("sinon");
import nedb = require("nedb");
import express = require("express");
import winston = require("winston");
import speakeasy = require("speakeasy");
import u2f = require("u2f");
import nodemailer = require("nodemailer");
import session = require("express-session");

import { AppConfiguration, UserConfiguration } from "../../src/types/Configuration";
import { GlobalDependencies, Nodemailer } from "../../src/types/Dependencies";
import Server from "../../src/server/lib/Server";


describe("test server configuration", function () {
  let deps: GlobalDependencies;
  let sessionMock: sinon.SinonSpy;

  before(function () {
    const transporter = {
      sendMail: sinon.stub().yields()
    };

    const createTransport = sinon.stub(nodemailer, "createTransport");
    createTransport.returns(transporter);

    sessionMock = sinon.spy(session);

    deps = {
      nodemailer: nodemailer,
      speakeasy: speakeasy,
      u2f: u2f,
      nedb: nedb,
      winston: winston,
      ldapjs: {
        createClient: sinon.spy(function () {
          return { on: sinon.spy() };
        })
      },
      session: sessionMock as any
    };
  });


  it("should set cookie scope to domain set in the config", function () {
    const config = {
      session: {
        domain: "example.com",
        secret: "secret"
      },
      ldap: {
        url: "http://ldap",
        user: "user",
        password: "password"
      },
      notifier: {
        gmail: {
          username: "user@example.com",
          password: "password"
        }
      }
    } as UserConfiguration;

    const server = new Server();
    server.start(config, deps);

    assert(sessionMock.calledOnce);
    assert.equal(sessionMock.getCall(0).args[0].cookie.domain, "example.com");
  });
});
