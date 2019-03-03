
import { INotifier } from "../notifiers/INotifier";

import Fs = require("fs");
import Path = require("path");
import Ejs = require("ejs");
import BluebirdPromise = require("bluebird");

const email_template = Fs.readFileSync(Path.join(__dirname, "../../resources/email-template.ejs"), "UTF-8");

export abstract class AbstractEmailNotifier implements INotifier {
  notify(to: string, subject: string, link: string): BluebirdPromise<void> {
    const d = {
      url: link,
      button_title: "Continue",
      title: subject
    };
    return this.sendEmail(to, subject, Ejs.render(email_template, d));
  }

  abstract sendEmail(to: string, subject: string, content: string): BluebirdPromise<void>;
}