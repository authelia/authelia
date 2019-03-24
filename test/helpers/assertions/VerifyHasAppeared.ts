import SeleniumWebDriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, locator: SeleniumWebDriver.Locator, timeout: number = 5000) {
  await driver.wait(SeleniumWebDriver.until.elementLocated(locator), timeout);
}