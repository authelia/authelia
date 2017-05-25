
import * as sinon from "sinon";
import * as assert from "assert";
import { FileSystemNotifier } from "../../../src/server/lib/notifiers/FileSystemNotifier";
import * as tmp from "tmp";
import * as fs from "fs";
import BluebirdPromise = require("bluebird");

const NOTIFICATIONS_DIRECTORY = "notifications";

describe("test FS notifier", function() {
  let tmpDir: tmp.SynchrounousResult;
  before(function() {
    tmpDir = tmp.dirSync({ unsafeCleanup: true });
  });

  after(function() {
    tmpDir.removeCallback();
  });

  it("should write the notification in a file", function() {
    const options = {
      filename: tmpDir.name + "/" + NOTIFICATIONS_DIRECTORY
    };

    const sender = new FileSystemNotifier(options);
    const subject = "subject";

    const identity = {
      userid: "user",
      email: "user@example.com"
    };

    const url = "http://test.com";

    return sender.notify(identity, subject, url)
    .then(function() {
      const content = fs.readFileSync(options.filename, "UTF-8");
      assert(content.length > 0);
      return BluebirdPromise.resolve();
    });
  });
});
