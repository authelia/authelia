
import { INotifier } from "../notifiers/INotifier";
import { Identity } from "../../../types/Identity";

import Fs = require("fs");
import Path = require("path");
import Ejs = require("ejs");
import BluebirdPromise = require("bluebird");

const email_template = Fs.readFileSync(Path.join(__dirname, "../../resources/email-template.ejs"), "UTF-8");

export abstract class AbstractEmailNotifier implements INotifier {

  notify(identity: Identity, subject: string, link: string): BluebirdPromise<void> {
    const d = {
      url: link,
      button_title: "Continue",
      title: subject
    };
    return this.sendEmail(identity.email, subject, Ejs.render(email_template, d));
  }

  abstract sendEmail(email: string, subject: string, content: string): BluebirdPromise<void>;
}