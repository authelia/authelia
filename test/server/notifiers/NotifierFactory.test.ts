
import * as sinon from "sinon";
import * as BluebirdPromise from "bluebird";
import * as assert from "assert";

import { NotifierFactory } from "../../../src/server/lib/notifiers/NotifierFactory";
import { GMailNotifier } from "../../../src/server/lib/notifiers/GMailNotifier";
import { FileSystemNotifier } from "../../../src/server/lib/notifiers/FileSystemNotifier";

import NodemailerMock = require("../mocks/nodemailer");


describe("test notifier factory", function() {
  let nodemailerMock: NodemailerMock.NodemailerMock;
  it("should build a Gmail Notifier", function() {
    const options = {
      gmail: {
        username: "abc",
        password: "password"
      }
    };
    nodemailerMock = NodemailerMock.NodemailerMock();
    nodemailerMock.createTransport.returns(sinon.spy());
    assert(NotifierFactory.build(options, nodemailerMock) instanceof GMailNotifier);
  });

  it("should build a FS Notifier", function() {
    const options = {
      filesystem: {
        filename: "abc"
      }
    };

    assert(NotifierFactory.build(options, nodemailerMock) instanceof FileSystemNotifier);
  });
});
