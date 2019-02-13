import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";
import Assert = require("assert");

export default async function(driver: WebDriver, message: string, timeout: number = 5000) {
  await driver.wait(SeleniumWebdriver.until.elementLocated(
    SeleniumWebdriver.By.className("notification")), timeout)
  const notificationEl = driver.findElement(SeleniumWebdriver.By.className("notification"));
  const txt = await notificationEl.getText();
  Assert.equal(message, txt);
}