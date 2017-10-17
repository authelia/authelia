import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  When("I sleep for {number} seconds", function (seconds: number) {
    return this.driver.sleep(seconds * 1000);
  });
});