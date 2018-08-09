import Speakeasy = require("speakeasy");
import SeleniumWebdriver = require("selenium-webdriver");
import ClickOnButton from "./click-on-button";

export default function(driver: any, secret: string) {
  const token = Speakeasy.totp({
    secret: secret,
    encoding: "base32"
  });
  return driver.wait(
    SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.id("token")), 5000)
      .then(function () {
        return driver.findElement(SeleniumWebdriver.By.id("token"))
          .sendKeys(token);
      })
      .then(function () {
        return ClickOnButton(driver, "Sign in");
      });
}