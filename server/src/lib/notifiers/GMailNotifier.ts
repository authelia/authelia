
import * as BluebirdPromise from "bluebird";

import { AbstractEmailNotifier } from "../notifiers/AbstractEmailNotifier";
import { GmailNotifierConfiguration } from "../configuration/Configuration";
import { IMailSender } from "./IMailSender";

export class GMailNotifier extends AbstractEmailNotifier {
  private mailSender: IMailSender;

  constructor(options: GmailNotifierConfiguration, mailSender: IMailSender) {
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
    return this.mailSender.send(mailOptions);
  }
}
