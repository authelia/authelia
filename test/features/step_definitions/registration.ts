import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  When("the otpauth url has label {stringInDoubleQuotes} and issuer \
{stringInDoubleQuotes}", function (label: string, issuer: string) {
      return this.driver.findElement(seleniumWebdriver.By.id("qrcode"))
        .getAttribute("title")
        .then(function (title: string) {
          const re = `^otpauth://totp/${label}\\?secret=[A-Z0-9]+&issuer=${issuer}$`;
          Assert(new RegExp(re).test(title));
        })
    });
});
