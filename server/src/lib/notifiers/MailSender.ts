import { IMailSender } from "./IMailSender";
import Nodemailer = require("nodemailer");
import NodemailerDirectTransport = require("nodemailer-direct-transport");
import NodemailerSmtpTransport = require("nodemailer-smtp-transport");
import BluebirdPromise = require("bluebird");

export class MailSender implements IMailSender {
  private transporter: Nodemailer.Transporter;

  constructor(options: NodemailerDirectTransport.DirectOptions |
    NodemailerSmtpTransport.SmtpOptions, nodemailer: typeof Nodemailer) {
    this.transporter = nodemailer.createTransport(options);
  }

  verify(): BluebirdPromise<void> {
    const that = this;
    return new BluebirdPromise(function (resolve, reject) {
      that.transporter.verify(function (error: Error, success: any) {
        if (error) {
          reject(new Error("Unable to connect to SMTP server. \
  Please check the service is running and your credentials are correct."));
          return;
        }
        resolve();
      });
    });
  }

  send(mailOptions: Nodemailer.SendMailOptions): BluebirdPromise<void> {
    const that = this;
    return new BluebirdPromise(function (resolve, reject) {
      that.transporter.sendMail(mailOptions, (error: Error,
        data: Nodemailer.SentMessageInfo) => {
        if (error) {
          reject(new Error("Error while sending email: " + error.message));
          return;
        }
        resolve();
      });
    });
  }
}