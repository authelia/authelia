import SeleniumWebdriver = require("selenium-webdriver");

export default function(driver: any, fieldName: string, text: string) {
  return driver.wait(
    SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.name(fieldName)), 5000)
    .then(function (el) {
      return el.sendKeys(text);
    });
};