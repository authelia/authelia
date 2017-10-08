import * as sinon from "sinon";
import * as assert from "assert";
import BluebirdPromise = require("bluebird");

import { MailSenderStub } from "../mocks/notifiers/MailSenderStub";
import GMailNotifier = require("../../src/lib/notifiers/GMailNotifier");


describe("test gmail notifier", function () {
  it("should send an email to given user", function () {
    const mailSender = new MailSenderStub();
    const options = {
      username: "user_gmail",
      password: "pass_gmail"
    };

    mailSender.sendStub.returns(BluebirdPromise.resolve());
    const sender = new GMailNotifier.GMailNotifier(options, mailSender);
    const subject = "subject";

    const identity = {
      userid: "user",
      email: "user@example.com"
    };

    const url = "http://test.com";

    return sender.notify(identity, subject, url)
      .then(function () {
        assert.equal(mailSender.sendStub.getCall(0).args[0].to, "user@example.com");
        assert.equal(mailSender.sendStub.getCall(0).args[0].subject, "subject");
        return BluebirdPromise.resolve();
      });
  });

  it("should fail while sending an email", function () {
    const mailSender = new MailSenderStub();
    const options = {
      username: "user_gmail",
      password: "pass_gmail"
    };

    mailSender.sendStub.returns(BluebirdPromise.reject(new Error("Failed to send mail")));
    const sender = new GMailNotifier.GMailNotifier(options, mailSender);
    const subject = "subject";

    const identity = {
      userid: "user",
      email: "user@example.com"
    };

    const url = "http://test.com";

    return sender.notify(identity, subject, url)
      .then(function () {
        return BluebirdPromise.reject(new Error());
      }, function() {
        return BluebirdPromise.resolve();
      });
  });
});
