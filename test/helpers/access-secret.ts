import SeleniumWebdriver = require("selenium-webdriver");
import Bluebird = require("bluebird");

export default function(driver: any) {
  return driver.findElement(
    SeleniumWebdriver.By.tagName('h1')).getText()
    .then(function(content: string) {
      return (content.indexOf('Secret') > -1) 
        ? Bluebird.resolve()
        : Bluebird.reject(new Error("Secret is not accessible."));
    })
}