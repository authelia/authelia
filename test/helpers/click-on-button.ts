import SeleniumWebdriver = require("selenium-webdriver");

export default function(driver: any, buttonText: string) {
  return driver.wait(
    SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.tagName("button")), 5000)
    .then(function () {
      return driver
        .findElement(SeleniumWebdriver.By.tagName("button"))
        .findElement(SeleniumWebdriver.By.xpath("//button[contains(.,'" + buttonText + "')]"))
        .click();
    });
};