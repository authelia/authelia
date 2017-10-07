import { IMailSenderBuilder } from "../../../src/lib/notifiers/IMailSenderBuilder";
import BluebirdPromise = require("bluebird");
import Nodemailer = require("nodemailer");
import Sinon = require("sinon");
import { IMailSender } from "../../../src/lib/notifiers/IMailSender";
import { SmtpNotifierConfiguration, GmailNotifierConfiguration } from "../../../src/lib/configuration/Configuration";

export class MailSenderBuilderStub implements IMailSenderBuilder {
  buildGmailStub: Sinon.SinonStub;
  buildSmtpStub: Sinon.SinonStub;

  constructor() {
    this.buildGmailStub = Sinon.stub();
    this.buildSmtpStub = Sinon.stub();
  }

  buildGmail(options: GmailNotifierConfiguration): IMailSender {
    return this.buildGmailStub(options);
  }

  buildSmtp(options: SmtpNotifierConfiguration): IMailSender {
    return this.buildSmtpStub(options);
  }

}