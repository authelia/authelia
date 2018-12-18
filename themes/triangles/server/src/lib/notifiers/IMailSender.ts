import BluebirdPromise = require("bluebird");
import Nodemailer = require("nodemailer");

export interface IMailSender {
  send(mailOptions: Nodemailer.SendMailOptions): BluebirdPromise<void>;
}