

import * as BluebirdPromise from "bluebird";

import { IMailSender } from "./IMailSender";
import { AbstractEmailNotifier } from "../notifiers/AbstractEmailNotifier";
import { SmtpNotifierConfiguration } from "../configuration/Configuration";

export class SmtpNotifier extends AbstractEmailNotifier {
  private mailSender: IMailSender;

  constructor(options: SmtpNotifierConfiguration,
    mailSender: IMailSender) {
    super();
    this.mailSender = mailSender;
  }

  sendEmail(email: string, subject: string, content: string) {
    const mailOptions = {
      from: "authelia@authelia.com",
      to: email,
      subject: subject,
      html: content
    };
    const that = this;
    return this.mailSender.send(mailOptions);
  }
}
