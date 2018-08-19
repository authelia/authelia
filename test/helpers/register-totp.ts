import Bluebird = require("bluebird");
import SeleniumWebdriver = require("selenium-webdriver");
import Fs = require("fs");
import Request = require("request-promise");

function retrieveValidationLinkFromNotificationFile(): Bluebird<string> {
  return Bluebird.promisify(Fs.readFile)("/tmp/authelia/notification.txt")
    .then(function (data: any) {
      const regexp = new RegExp(/Link: (.+)/);
      const match = regexp.exec(data);
      const link = match[1];
      return Bluebird.resolve(link);
    });
};

function retrieveValidationLinkFromEmail(): Bluebird<string> {
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

export default function(driver: any, email?: boolean): Bluebird<string> {
  return driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.className("register-totp")), 5000)
    .then(function () {
      return driver.findElement(SeleniumWebdriver.By.className("register-totp")).click();
    })
    .then(function () {
      if(email) return retrieveValidationLinkFromEmail();
      else return retrieveValidationLinkFromNotificationFile();
    })
    .then(function (link: string) {
      return driver.get(link);
    })
    .then(function () {
      return driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.id("secret")), 5000);
    })
    .then(function () {
      return driver.findElement(SeleniumWebdriver.By.id("secret")).getText();
    });
};
