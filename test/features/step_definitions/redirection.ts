import {Given, When, Then} from "cucumber";
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");

Given("I'm on {string}", function (link: string) {
  return this.driver.get(link);
});

When("I click on the link to {string}", function (link: string) {
  return this.driver.findElement(seleniumWebdriver.By.linkText(link)).click();
});

Then("I'm redirected to {string}", function (link: string) {
  return this.waitUntilUrlContains(link);
});