
import { NotifierConfiguration } from "..//Configuration";
import { NotifierDependencies } from "../../types/Dependencies";
import { INotifier } from "./INotifier";

import { GMailNotifier } from "./GMailNotifier";
import { FileSystemNotifier } from "./FileSystemNotifier";

export class NotifierFactory {
  static build(options: NotifierConfiguration, deps: NotifierDependencies): INotifier {
    if ("gmail" in options) {
      return new GMailNotifier(options.gmail, deps.nodemailer);
    }
    else if ("filesystem" in options) {
      return new FileSystemNotifier(options.filesystem);
    }
  }
}




