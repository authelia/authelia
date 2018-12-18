import { IMailSender } from "./IMailSender";
import { SmtpNotifierConfiguration, EmailNotifierConfiguration } from "../configuration/schema/NotifierConfiguration";

export interface IMailSenderBuilder {
  buildEmail(options: EmailNotifierConfiguration): IMailSender;
  buildSmtp(options: SmtpNotifierConfiguration): IMailSender;
}