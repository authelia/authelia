import { IMailSenderBuilder } from "../../../src/lib/notifiers/IMailSenderBuilder";
import BluebirdPromise = require("bluebird");
import Nodemailer = require("nodemailer");
import Sinon = require("sinon");
import { IMailSender } from "../../../src/lib/notifiers/IMailSender";
import { SmtpNotifierConfiguration, EmailNotifierConfiguration } from "../../../src/lib/configuration/schema/NotifierConfiguration";

export class MailSenderBuilderStub implements IMailSenderBuilder {
  buildEmailStub: Sinon.SinonStub;
  buildSmtpStub: Sinon.SinonStub;

  constructor() {
    this.buildEmailStub = Sinon.stub();
    this.buildSmtpStub = Sinon.stub();
  }

  buildEmail(options: EmailNotifierConfiguration): IMailSender {
    return this.buildEmailStub(options);
  }

  buildSmtp(options: SmtpNotifierConfiguration): IMailSender {
    return this.buildSmtpStub(options);
  }

}