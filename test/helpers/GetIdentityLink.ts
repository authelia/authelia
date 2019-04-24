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
    uri: "https://mail.example.com:8080/messages",
    json: true,
    rejectUnauthorized: false,
  });
  const messageId = data[data.length - 1].id;
  const data2 = await Request({
    method: "GET",
    rejectUnauthorized: false,
    uri: `https://mail.example.com:8080/messages/${messageId}.html`
  });
  const regexp = new RegExp(/<a href="(.+)" class="button">.*<\/a>/);
  const match = regexp.exec(data2);
  if (match == null) {
    throw new Error('No match');
  }
  return match[1];
};