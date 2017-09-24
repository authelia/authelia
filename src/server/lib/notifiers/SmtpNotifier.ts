

import * as BluebirdPromise from "bluebird";
import Nodemailer = require("nodemailer");

import { AbstractEmailNotifier } from "../notifiers/AbstractEmailNotifier";
import { SmtpNotifierConfiguration } from "../configuration/Configuration";

export class SmtpNotifier extends AbstractEmailNotifier {
  private transporter: any;

  constructor(options: SmtpNotifierConfiguration, nodemailer: typeof Nodemailer) {
    super();
    const smtpOptions = {
      host: options.host,
      port: options.port,
      secure: options.secure, // upgrade later with STARTTLS
      auth: {
        user: options.username,
        pass: options.password
      }
    };
    console.log(smtpOptions);
    const transporter = nodemailer.createTransport(smtpOptions);
    this.transporter = BluebirdPromise.promisifyAll(transporter);

    // verify connection configuration
    console.log("Checking SMTP server connection.");
    transporter.verify(function (error, success) {
      if (error) {
        throw new Error("Unable to connect to SMTP server. \
Please check the service is running and your credentials are correct.");
      } else {
        console.log("SMTP Server is ready to take our messages");
      }
    });
  }

  sendEmail(email: string, subject: string, content: string) {
    const mailOptions = {
      from: "authelia@authelia.com",
      to: email,
      subject: subject,
      html: content
    };
    return this.transporter.sendMail(mailOptions, (error: Error, data: string) => {
      if (error) {
        return console.log(error);
      }
      console.log("Message sent: %s", JSON.stringify(data));
    });
  }
}
