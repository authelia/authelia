
import * as sinon from "sinon";
import * as BluebirdPromise from "bluebird";
import * as assert from "assert";

import { NotifierFactory } from "../../src/lib/notifiers/NotifierFactory";
import { GMailNotifier } from "../../src/lib/notifiers/GMailNotifier";
import { SmtpNotifier } from "../../src/lib/notifiers/SmtpNotifier";
import { MailSenderBuilderStub } from "../mocks/notifiers/MailSenderBuilderStub";


describe("test notifier factory", function () {
  let mailSenderBuilderStub: MailSenderBuilderStub;
  it("should build a Gmail Notifier", function () {
    const options = {
      gmail: {
        username: "abc",
        password: "password",
        sender: "admin@example.com"
      }
    };
    mailSenderBuilderStub = new MailSenderBuilderStub();
    assert(NotifierFactory.build(options, mailSenderBuilderStub) instanceof GMailNotifier);
  });

  it("should build a SMTP Notifier", function () {
    const options = {
      smtp: {
        username: "user",
        password: "pass",
        secure: true,
        host: "localhost",
        port: 25,
        sender: "admin@example.com"
      }
    };

    mailSenderBuilderStub = new MailSenderBuilderStub();
    assert(NotifierFactory.build(options, mailSenderBuilderStub) instanceof SmtpNotifier);
  });
});
