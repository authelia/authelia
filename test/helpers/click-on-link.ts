import SeleniumWebdriver = require("selenium-webdriver");

export default function(driver: any, linkText: string) {
  return driver.wait(
    SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.linkText(linkText)), 5000)
    .then(function (el) {
      return el.click();
    });
};