import { IMailSender } from "./IMailSender";
import { SmtpNotifierConfiguration, GmailNotifierConfiguration } from "../configuration/Configuration";

export interface IMailSenderBuilder {
  buildGmail(options: GmailNotifierConfiguration): IMailSender;
  buildSmtp(options: SmtpNotifierConfiguration): IMailSender;
}