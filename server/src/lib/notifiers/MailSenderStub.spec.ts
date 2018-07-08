import { IMailSender } from "../../../src/lib/notifiers/IMailSender";
import BluebirdPromise = require("bluebird");
import Nodemailer = require("nodemailer");
import Sinon = require("sinon");

export class MailSenderStub implements IMailSender {
  sendStub: Sinon.SinonStub;

  constructor() {
    this.sendStub = Sinon.stub();
  }

  send(mailOptions: Nodemailer.SendMailOptions): BluebirdPromise<void> {
    return this.sendStub(mailOptions);
  }
}