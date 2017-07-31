
import * as BluebirdPromise from "bluebird";
import * as util from "util";
import * as Fs from "fs";
import { INotifier } from "./INotifier";
import { Identity } from "../../../types/Identity";

import { FileSystemNotifierConfiguration } from "../configuration/Configuration";

export class FileSystemNotifier implements INotifier {
  private filename: string;

  constructor(options: FileSystemNotifierConfiguration) {
    this.filename = options.filename;
  }

  notify(identity: Identity, subject: string, link: string): BluebirdPromise<void> {
    const content = util.format("Date: %s\nUser: %s\nSubject: %s\nLink: %s", new Date().toString(), identity.userid,
      subject, link);
    const writeFilePromised: any = BluebirdPromise.promisify(Fs.writeFile);
    return writeFilePromised(this.filename, content);
  }
}

