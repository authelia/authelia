
import * as BluebirdPromise from "bluebird";
import * as fs from "fs";
import * as ejs from "ejs";
import nodemailer = require("nodemailer");

import { Nodemailer } from "../../../types/Dependencies";
import { Identity } from "../../../types/Identity";
import { INotifier } from "../notifiers/INotifier";
import { GmailNotifierConfiguration } from "../../../types/Configuration";
import path = require("path");

const email_template = fs.readFileSync(path.join(__dirname, "../../resources/email-template.ejs"), "UTF-8");

export class GMailNotifier extends INotifier {
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

  notify(identity: Identity, subject: string, link: string): BluebirdPromise<void> {
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
