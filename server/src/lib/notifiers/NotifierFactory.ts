
import { NotifierConfiguration } from "../configuration/Configuration";
import Nodemailer = require("nodemailer");
import { INotifier } from "./INotifier";

import { GMailNotifier } from "./GMailNotifier";
import { SmtpNotifier } from "./SmtpNotifier";
import { IMailSender } from "./IMailSender";
import { IMailSenderBuilder } from "./IMailSenderBuilder";

export class NotifierFactory {
  static build(options: NotifierConfiguration, mailSenderBuilder: IMailSenderBuilder): INotifier {
    if ("gmail" in options) {
      const mailSender = mailSenderBuilder.buildGmail(options.gmail);
      return new GMailNotifier(options.gmail, mailSender);
    }
    else if ("smtp" in options) {
      const mailSender = mailSenderBuilder.buildSmtp(options.smtp);
      return new SmtpNotifier(options.smtp, mailSender);
    }
    else {
      throw new Error("No available notifier option detected.");
    }
  }
}




