
import * as sinon from "sinon";
import * as BluebirdPromise from "bluebird";
import * as assert from "assert";

import NodemailerMock = require("../mocks/nodemailer");

import { NotifierFactory } from "../../../src/lib/notifiers/NotifierFactory";
import { GMailNotifier } from "../../../src/lib/notifiers/GMailNotifier";
import { FileSystemNotifier } from "../../../src/lib/notifiers/FileSystemNotifier";

import { NotifierDependencies } from "../../../src/types/Dependencies";


describe("test notifier", function() {
  const deps: NotifierDependencies = {
    nodemailer: NodemailerMock
  };

  it("should build a Gmail Notifier", function() {
    const options = {
      gmail: {
        username: "abc",
        password: "password"
      }
    };
    assert(NotifierFactory.build(options, deps) instanceof GMailNotifier);
  });

  it("should build a FS Notifier", function() {
    const options = {
      filesystem: {
        filename: "abc"
      }
    };

    assert(NotifierFactory.build(options, deps) instanceof FileSystemNotifier);
  });
});
