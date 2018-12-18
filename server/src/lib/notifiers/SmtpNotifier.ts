

import * as BluebirdPromise from "bluebird";

import { IMailSender } from "./IMailSender";
import { AbstractEmailNotifier } from "../notifiers/AbstractEmailNotifier";
import { SmtpNotifierConfiguration } from "../configuration/schema/NotifierConfiguration";

export class SmtpNotifier extends AbstractEmailNotifier {
  private mailSender: IMailSender;
  private sender: string;

  constructor(options: SmtpNotifierConfiguration,
    mailSender: IMailSender) {
    super();
    this.mailSender = mailSender;
    this.sender = options.sender;
  }

  sendEmail(to: string, subject: string, content: string) {
    const mailOptions = {
      from: this.sender,
      to: to,
      subject: subject,
      html: content
    };
    const that = this;
    return this.mailSender.send(mailOptions);
  }
}
