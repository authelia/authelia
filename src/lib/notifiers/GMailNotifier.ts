
import * as Promise from "bluebird";
import * as fs from "fs";
import * as ejs from "ejs";
import nodemailer = require("nodemailer");

import { NodemailerDependencies } from "../Dependencies";
import { Identity } from "../Identity";
import { INotifier } from "../notifiers/INotifier";
import { GmailNotifierConfiguration } from "../Configuration";

const email_template = fs.readFileSync(__dirname + "/../../resources/email-template.ejs", "UTF-8");

export class GMailNotifier extends INotifier {
  private transporter: any;

  constructor(options: GmailNotifierConfiguration, deps: NodemailerDependencies) {
    super();
    const transporter = deps.createTransport({
      service: "gmail",
      auth: {
        user: options.username,
        pass: options.password
      }
    });
    this.transporter = Promise.promisifyAll(transporter);
  }

  notify(identity: Identity, subject: string, link: string): Promise<void> {
    const d = {
      url: link,
      button_title: "Continue",
      title: subject
    };

    const mailOptions = {
      from: "auth-server@open-intent.io",
      to: identity.email,
      subject: subject,
      html: ejs.render(email_template, d)
    };
    return this.transporter.sendMailAsync(mailOptions);
  }
}
