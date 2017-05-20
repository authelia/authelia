
import * as sinon from "sinon";
import * as BluebirdPromise from "bluebird";
import * as assert from "assert";

import NodemailerMock = require("../mocks/nodemailer");

import { NotifierFactory } from "../../../src/lib/notifiers/NotifierFactory";
import { GMailNotifier } from "../../../src/lib/notifiers/GMailNotifier";
import { FileSystemNotifier } from "../../../src/lib/notifiers/FileSystemNotifier";

import nodemailerMock = require("../mocks/nodemailer");


describe("test notifier factory", function() {
  it("should build a Gmail Notifier", function() {
    const options = {
      gmail: {
        username: "abc",
        password: "password"
      }
    };
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
