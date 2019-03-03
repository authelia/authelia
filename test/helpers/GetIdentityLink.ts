import Bluebird = require("bluebird");
import Fs = require("fs");
import Request = require("request-promise");

export async function GetLinkFromFile() {
  const data = await Bluebird.promisify(Fs.readFile)("/tmp/authelia/notification.txt")
  const regexp = new RegExp(/Link: (.+)/);
  const match = regexp.exec(data.toLocaleString());
  if (match == null) {
    throw new Error('No match');
  }
  return match[1];
};

export async function GetLinkFromEmail() {
  const data = await Request({
    method: "GET",
    uri: "http://localhost:8085/messages",
    json: true
  });
  const messageId = data[data.length - 1].id;
  const data2 = await Request({
    method: "GET",
    uri: `http://localhost:8085/messages/${messageId}.html`
  });
  const regexp = new RegExp(/<a href="(.+)" class="button">Continue<\/a>/);
  const match = regexp.exec(data2);
  if (match == null) {
    throw new Error('No match');
  }
  return match[1];
};