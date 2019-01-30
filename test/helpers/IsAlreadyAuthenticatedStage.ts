import SeleniumWebDriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver) {
  await driver.wait(SeleniumWebDriver.until.elementLocated(SeleniumWebDriver.By.className('already-authenticated-step')));
}