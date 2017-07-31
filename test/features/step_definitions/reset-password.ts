import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");
import Fs = require("fs");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  When("I click on the link {stringInDoubleQuotes}", function (text: string) {
    return this.driver.findElement(seleniumWebdriver.By.linkText(text)).click();
  });

  When("I click on the link of the email", function () {
    const notif = Fs.readFileSync("./notifications/notification.txt").toString();
    const regexp = new RegExp(/Link: (.+)/);
    const match = regexp.exec(notif);
    const link = match[1];
    const that = this;

    return this.driver.get(link);
  });
});