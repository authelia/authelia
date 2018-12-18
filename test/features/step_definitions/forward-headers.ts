import {Then} from "cucumber";
import seleniumWebdriver = require("selenium-webdriver");
import CustomWorld = require("../support/world");
import Util = require("util");
import Bluebird = require("bluebird");
import Request = require("request-promise");

Then("I see header {string} set to {string}",
  { timeout: 5000 },
  function (expectedHeaderName: string, expectedValue: string) {
    return this.driver.findElement(seleniumWebdriver.By.tagName("body")).getText()
      .then(function (txt: string) {
        const expectedLine = Util.format("\"%s\": \"%s\"", expectedHeaderName, expectedValue);
        if (txt.indexOf(expectedLine) > 0)
          return Bluebird.resolve();
        else
          return Bluebird.reject(new Error(Util.format("No such header or with unexpected value.")));
      });
  })