import Bluebird = require("bluebird");
import SeleniumWebdriver = require("selenium-webdriver");
import Fs = require("fs");

function retrieveValidationLinkFromNotificationFile(): Bluebird<string> {
  return Bluebird.promisify(Fs.readFile)("/tmp/authelia/notification.txt")
    .then(function (data: any) {
      const regexp = new RegExp(/Link: (.+)/);
      const match = regexp.exec(data);
      const link = match[1];
      return Bluebird.resolve(link);
    });
};

export default function(driver: any): Bluebird<string> {
  return driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.className("register-totp")), 5000)
    .then(function () {
      return driver.findElement(SeleniumWebdriver.By.className("register-totp")).click();
    })
    .then(function () {
      return retrieveValidationLinkFromNotificationFile();
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
