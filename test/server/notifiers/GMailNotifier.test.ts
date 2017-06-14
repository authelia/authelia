import * as sinon from "sinon";
import * as assert from "assert";
import BluebirdPromise = require("bluebird");

import NodemailerMock = require("../mocks/nodemailer");
import GMailNotifier = require("../../../src/server/lib/notifiers/GMailNotifier");


describe("test gmail notifier", function () {
  it("should send an email", function () {
    const transporter = {
      sendMail: sinon.stub().yields()
    };
    const nodemailerMock = NodemailerMock.NodemailerMock();
    nodemailerMock.createTransport.returns(transporter);

    const options = {
      username: "user_gmail",
      password: "pass_gmail"
    };

    const sender = new GMailNotifier.GMailNotifier(options, nodemailerMock);
    const subject = "subject";

    const identity = {
      userid: "user",
      email: "user@example.com"
    };

    const url = "http://test.com";

    return sender.notify(identity, subject, url)
      .then(function () {
        assert.equal(nodemailerMock.createTransport.getCall(0).args[0].auth.user, "user_gmail");
        assert.equal(nodemailerMock.createTransport.getCall(0).args[0].auth.pass, "pass_gmail");
        assert.equal(transporter.sendMail.getCall(0).args[0].to, "user@example.com");
        assert.equal(transporter.sendMail.getCall(0).args[0].subject, "subject");
        return BluebirdPromise.resolve();
      });
  });
});
