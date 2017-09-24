
import * as sinon from "sinon";
import * as BluebirdPromise from "bluebird";
import * as assert from "assert";

import { NotifierFactory } from "../../../../src/server/lib/notifiers/NotifierFactory";
import { GMailNotifier } from "../../../../src/server/lib/notifiers/GMailNotifier";
import { SmtpNotifier } from "../../../../src/server/lib/notifiers/SmtpNotifier";

import NodemailerMock = require("../mocks/nodemailer");


describe("test notifier factory", function() {
  let nodemailerMock: NodemailerMock.NodemailerMock;
  it("should build a Gmail Notifier", function() {
    const options = {
      gmail: {
        username: "abc",
        password: "password"
      }
    };
    nodemailerMock = NodemailerMock.NodemailerMock();
    const transporterMock = NodemailerMock.NodemailerTransporterMock();
    nodemailerMock.createTransport.returns(transporterMock);
    assert(NotifierFactory.build(options, nodemailerMock) instanceof GMailNotifier);
  });

  it("should build a SMTP Notifier", function() {
    const options = {
      smtp: {
        username: "user",
        password: "pass",
        secure: true,
        host: "localhost",
        port: 25
      }
    };

    nodemailerMock = NodemailerMock.NodemailerMock();
    const transporterMock = NodemailerMock.NodemailerTransporterMock();
    nodemailerMock.createTransport.returns(transporterMock);
    assert(NotifierFactory.build(options, nodemailerMock) instanceof SmtpNotifier);
  });
});
