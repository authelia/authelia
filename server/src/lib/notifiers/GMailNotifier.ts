
import * as BluebirdPromise from "bluebird";
import nodemailer = require("nodemailer");

import { Nodemailer } from "../../../types/Dependencies";
import { AbstractEmailNotifier } from "../notifiers/AbstractEmailNotifier";
import { GmailNotifierConfiguration } from "../configuration/Configuration";

export class GMailNotifier extends AbstractEmailNotifier {
  private transporter: any;

  constructor(options: GmailNotifierConfiguration, nodemailer: Nodemailer) {
    super();
    const transporter = nodemailer.createTransport({
      service: "gmail",
      auth: {
        user: options.username,
        pass: options.password
      }
    });
    this.transporter = BluebirdPromise.promisifyAll(transporter);
  }

  sendEmail(email: string, subject: string, content: string) {
    const mailOptions = {
      from: "authelia@authelia.com",
      to: email,
      subject: subject,
      html: content
    };
    return this.transporter.sendMailAsync(mailOptions);
  }
}
