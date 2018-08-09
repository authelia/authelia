import SeleniumWebdriver = require("selenium-webdriver");

export default function(driver: any, url: string, timeout: number = 5000) {
  return driver.get(url)
    .then(function () {
      return driver.wait(SeleniumWebdriver.until.urlIs(url), timeout);
    });
}