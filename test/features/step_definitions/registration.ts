import {When} from "cucumber";
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");

When("the otpauth url has label {string} and issuer \
{string}", function (label: string, issuer: string) {
    return this.driver.findElement(seleniumWebdriver.By.id("qrcode"))
      .getAttribute("title")
      .then(function (title: string) {
        const re = `^otpauth://totp/${label}\\?secret=[A-Z0-9]+&issuer=${issuer}$`;
        Assert(new RegExp(re).test(title));
      })
  });
