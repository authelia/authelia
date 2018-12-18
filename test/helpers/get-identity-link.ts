import Bluebird = require("bluebird");
import Fs = require("fs");
import Request = require("request-promise");

export function GetLinkFromFile(): Bluebird<string> {
  return Bluebird.promisify(Fs.readFile)("/tmp/authelia/notification.txt")
    .then(function (data: any) {
      const regexp = new RegExp(/Link: (.+)/);
      const match = regexp.exec(data);
      const link = match[1];
      return Bluebird.resolve(link);
    });
};

export function GetLinkFromEmail(): Bluebird<string> {
  return Request({
    method: "GET",
    uri: "http://localhost:8085/messages",
    json: true
  })
    .then(function (data: any) {
      const messageId = data[data.length - 1].id;
      return Request({
        method: "GET",
        uri: `http://localhost:8085/messages/${messageId}.html`
      });
    })
    .then(function (data: any) {
      const regexp = new RegExp(/<a href="(.+)" class="button">Continue<\/a>/);
      const match = regexp.exec(data);
      const link = match[1];
      return Bluebird.resolve(link);
    });
};