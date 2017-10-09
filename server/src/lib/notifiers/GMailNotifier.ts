
import * as BluebirdPromise from "bluebird";

import { AbstractEmailNotifier } from "../notifiers/AbstractEmailNotifier";
import { GmailNotifierConfiguration } from "../configuration/Configuration";
import { IMailSender } from "./IMailSender";

export class GMailNotifier extends AbstractEmailNotifier {
  private mailSender: IMailSender;
  private sender: string;

  constructor(options: GmailNotifierConfiguration, mailSender: IMailSender) {
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
    return this.mailSender.send(mailOptions);
  }
}
