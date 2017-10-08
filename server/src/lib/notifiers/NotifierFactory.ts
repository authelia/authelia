
import { NotifierConfiguration } from "../configuration/Configuration";
import { Nodemailer } from "../../../types/Dependencies";
import { INotifier } from "./INotifier";

import { FileSystemNotifier } from "./FileSystemNotifier";
import { GMailNotifier } from "./GMailNotifier";
import { SmtpNotifier } from "./SmtpNotifier";

export class NotifierFactory {
  static build(options: NotifierConfiguration, nodemailer: Nodemailer): INotifier {
    if ("gmail" in options) {
      return new GMailNotifier(options.gmail, nodemailer);
    }
    else if ("smtp" in options) {
      return new SmtpNotifier(options.smtp, nodemailer);
    }
    else if ("filesystem" in options) {
      return new FileSystemNotifier(options.filesystem);
    }
    else {
      throw new Error("No available notifier option detected.");
    }
  }
}




