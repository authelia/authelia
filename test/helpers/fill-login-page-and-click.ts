import SeleniumWebdriver = require("selenium-webdriver");

export default function(driver: any, username: string, password: string) {
  return driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.id("username")), 5000)
    .then(function () {
      return driver.findElement(SeleniumWebdriver.By.id("username"))
        .sendKeys(username);
    })
    .then(function () {
      return driver.findElement(SeleniumWebdriver.By.id("password"))
        .sendKeys(password);
    })
    .then(function () {
      return driver.findElement(SeleniumWebdriver.By.tagName("button"))
        .click();
    });
};