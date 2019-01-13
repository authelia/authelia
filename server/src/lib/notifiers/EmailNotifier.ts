import { AbstractEmailNotifier } from "../notifiers/AbstractEmailNotifier";
import { EmailNotifierConfiguration } from "../configuration/schema/NotifierConfiguration";
import { IMailSender } from "./IMailSender";

export class EmailNotifier extends AbstractEmailNotifier {
  private mailSender: IMailSender;
  private sender: string;

  constructor(options: EmailNotifierConfiguration, mailSender: IMailSender) {
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
