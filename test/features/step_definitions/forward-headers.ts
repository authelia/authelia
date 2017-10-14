import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import CustomWorld = require("../support/world");
import Util = require("util");
import BluebirdPromise = require("bluebird");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  Then("I see header {stringInDoubleQuotes} set to {stringInDoubleQuotes}",
    { timeout: 5000 },
    function (expectedHeaderName: string, expectedValue: string) {
      return this.driver.findElement(seleniumWebdriver.By.tagName("body")).getText()
        .then(function (txt: string) {
          const expectedLine = Util.format("\"%s\": \"%s\"", expectedHeaderName, expectedValue);          
          if (txt.indexOf(expectedLine) > 0)
            return BluebirdPromise.resolve();
          else
            return BluebirdPromise.reject(new Error(Util.format("No such header or with unexpected value.")));
        });
    })
});