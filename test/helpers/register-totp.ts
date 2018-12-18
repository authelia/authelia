import Bluebird = require("bluebird");
import SeleniumWebdriver = require("selenium-webdriver");
import {GetLinkFromFile, GetLinkFromEmail} from '../helpers/get-identity-link';

export default function(driver: any, email?: boolean): Bluebird<string> {
  return driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.className("register-totp")), 5000)
    .then(function () {
      return driver.findElement(SeleniumWebdriver.By.className("register-totp")).click();
    })
    .then(function () {
      if(email) return GetLinkFromEmail();
      else return GetLinkFromFile();
    })
    .then(function (link: string) {
      return driver.get(link);
    })
    .then(function () {
      return driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.id("secret")), 5000);
    })
    .then(function () {
      return driver.findElement(SeleniumWebdriver.By.id("secret")).getText();
    });
};
