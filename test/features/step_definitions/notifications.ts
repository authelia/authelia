import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");
import Fs = require("fs");
import CustomWorld = require("../support/world");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  Then("I get a notification of type {stringInDoubleQuotes} with message {stringInDoubleQuotes}",
    function (notificationType: string, notificationMessage: string) {
      const that = this;
      const notificationEl = this.driver.findElement(seleniumWebdriver.By.className("notification"));
      return this.driver.wait(seleniumWebdriver.until.elementIsVisible(notificationEl), 2000)
        .then(function () {
          return notificationEl.getText();
        })
        .then(function (txt: string) {
          Assert.equal(notificationMessage, txt);
          return notificationEl.getAttribute("class");
        })
        .then(function(classes: string) {
          Assert(classes.indexOf(notificationType) > -1, "Class '" + notificationType + "' not found in notification element.");
        });
    });

});