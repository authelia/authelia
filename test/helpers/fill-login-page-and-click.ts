import Bluebird = require("bluebird");
import SeleniumWebdriver = require("selenium-webdriver");

export default function(
  driver: any,
  username: string,
  password: string,
  keepMeLoggedIn: boolean = false) {
  return driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.id("username")), 5000)
    .then(() => {
      return driver.findElement(SeleniumWebdriver.By.id("username"))
        .sendKeys(username);
    })
    .then(() => {
      return driver.findElement(SeleniumWebdriver.By.id("password"))
        .sendKeys(password);
    })
    .then(() => {
      if (keepMeLoggedIn) {
        return driver.findElement(SeleniumWebdriver.By.id("keep_me_logged_in"))
          .click();
      }
      return Bluebird.resolve();
    })
    .then(() => {
      return driver.findElement(SeleniumWebdriver.By.tagName("button"))
        .click();
    });
};