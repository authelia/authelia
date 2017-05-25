
import * as BluebirdPromise from "bluebird";
import * as util from "util";
import * as fs from "fs";
import { INotifier } from "./INotifier";
import { Identity } from "../../../types/Identity";

import { FileSystemNotifierConfiguration } from "../../../types/Configuration";

export class FileSystemNotifier extends INotifier {
  private filename: string;

  constructor(options: FileSystemNotifierConfiguration) {
    super();
    this.filename = options.filename;
  }

  notify(identity: Identity, subject: string, link: string): BluebirdPromise<void> {
    const content = util.format("Date: %s\nUser: %s\nSubject: %s\nLink: %s", new Date().toString(), identity.userid,
      subject, link);
    const writeFilePromised = BluebirdPromise.promisify<void, string, string>(fs.writeFile);
    return writeFilePromised(this.filename, content);
  }
}

