
import { NotifierConfiguration } from "../../../types/Configuration";
import { Nodemailer } from "../../../types/Dependencies";
import { INotifier } from "./INotifier";

import { GMailNotifier } from "./GMailNotifier";
import { FileSystemNotifier } from "./FileSystemNotifier";

export class NotifierFactory {
  static build(options: NotifierConfiguration, nodemailer: Nodemailer): INotifier {
    if ("gmail" in options) {
      return new GMailNotifier(options.gmail, nodemailer);
    }
    else if ("filesystem" in options) {
      return new FileSystemNotifier(options.filesystem);
    }
  }
}




