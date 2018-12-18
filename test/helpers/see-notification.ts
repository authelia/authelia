import SeleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");

export default function(driver: any, type: string, message: string) {
  const notificationEl = driver.findElement(SeleniumWebdriver.By.className("notification"));
  return driver.wait(SeleniumWebdriver.until.elementIsVisible(notificationEl), 5000)
    .then(function () {
      return notificationEl.getText();
    })
    .then(function (txt: string) {
      Assert.equal(message, txt);
      return notificationEl.getAttribute("class");
    })
    .then(function (classes: string) {
      Assert(classes.indexOf(type) > -1, "Class '" + type + "' not found in notification element.");
      return driver.sleep(500);
    });
}