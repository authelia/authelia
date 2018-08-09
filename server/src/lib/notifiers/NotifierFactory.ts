
import { NotifierConfiguration } from "../configuration/schema/NotifierConfiguration";
import Nodemailer = require("nodemailer");
import { INotifier } from "./INotifier";

import { FileSystemNotifier } from "./FileSystemNotifier";
import { EmailNotifier } from "./EmailNotifier";
import { SmtpNotifier } from "./SmtpNotifier";
import { IMailSender } from "./IMailSender";
import { IMailSenderBuilder } from "./IMailSenderBuilder";

export class NotifierFactory {
  static build(options: NotifierConfiguration, mailSenderBuilder: IMailSenderBuilder): INotifier {
    if ("email" in options) {
      const mailSender = mailSenderBuilder.buildEmail(options.email);
      return new EmailNotifier(options.email, mailSender);
    }
    else if ("smtp" in options) {
      const mailSender = mailSenderBuilder.buildSmtp(options.smtp);
      return new SmtpNotifier(options.smtp, mailSender);
    }
    else if ("filesystem" in options) {
      return new FileSystemNotifier(options.filesystem);
    }
    else {
      throw new Error("No available notifier option detected.");
    }
  }
}




