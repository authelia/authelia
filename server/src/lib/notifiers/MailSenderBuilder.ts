import { IMailSender } from "./IMailSender";
import { IMailSenderBuilder } from "./IMailSenderBuilder";
import { MailSender } from "./MailSender";
import Nodemailer = require("nodemailer");
import NodemailerSmtpTransport = require("nodemailer-smtp-transport");
import { SmtpNotifierConfiguration, EmailNotifierConfiguration } from "../configuration/schema/NotifierConfiguration";

export class MailSenderBuilder implements IMailSenderBuilder {
  private nodemailer: typeof Nodemailer;

  constructor(nodemailer: typeof Nodemailer) {
    this.nodemailer = nodemailer;
  }

  buildEmail(options: EmailNotifierConfiguration): IMailSender {
    const emailOptions = {
      service: options.service,
      auth: {
        user: options.username,
        pass: options.password
      }
    };
    return new MailSender(emailOptions, this.nodemailer);
  }

  buildSmtp(options: SmtpNotifierConfiguration): IMailSender {
    const smtpOptions: NodemailerSmtpTransport.SmtpOptions = {
      host: options.host,
      port: options.port,
      secure: options.secure, // upgrade later with STARTTLS
    };

    if (options.username && options.password) {
      smtpOptions.auth = {
        user: options.username,
        pass: options.password
      };
    }

    return new MailSender(smtpOptions, this.nodemailer);
  }
}