import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  Given("I'm on https://{string}", function (link: string) {
    return this.driver.get("https://" + link);
  });

  When("I click on the link to {string}", function (link: string) {
    return this.driver.findElement(seleniumWebdriver.By.linkText(link)).click();
  });

  Then("I'm redirected to {stringInDoubleQuotes}", function (link: string) {
    return this.driver.wait(seleniumWebdriver.until.urlContains(link), 5000);
  });
});