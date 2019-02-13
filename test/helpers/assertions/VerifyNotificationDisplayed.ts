import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";
import Assert = require("assert");

export default async function(driver: WebDriver, message: string) {
  await driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.className("notification")), 5000)
  const notificationEl = driver.findElement(SeleniumWebdriver.By.className("notification"));
  const txt = await notificationEl.getText();
  Assert.equal(message, txt);
}