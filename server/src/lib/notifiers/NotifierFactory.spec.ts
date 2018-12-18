
import * as sinon from "sinon";
import * as BluebirdPromise from "bluebird";
import * as assert from "assert";

import { NotifierFactory } from "./NotifierFactory";
import { EmailNotifier } from "./EmailNotifier";
import { SmtpNotifier } from "./SmtpNotifier";
import { MailSenderBuilderStub } from "./MailSenderBuilderStub.spec";


describe("notifiers/NotifierFactory", function () {
  let mailSenderBuilderStub: MailSenderBuilderStub;
  it("should build a Email Notifier", function () {
    const options = {
      email: {
        username: "abc",
        password: "password",
        sender: "admin@example.com",
        service: "gmail"
      }
    };
    mailSenderBuilderStub = new MailSenderBuilderStub();
    assert(NotifierFactory.build(options, mailSenderBuilderStub) instanceof EmailNotifier);
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
