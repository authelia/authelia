import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");
import Fs = require("fs");
import CustomWorld = require("../support/world");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  When("I wait {number} seconds", { timeout: 10 * 1000 }, function (seconds: number) {
    return this.driver.sleep(seconds * 1000);
  });
});