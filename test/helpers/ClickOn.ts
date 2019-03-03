import SeleniumWebdriver, { WebDriver, Locator } from "selenium-webdriver";

export default async function(driver: WebDriver, locator: Locator, timeout: number = 5000) {
  const el = await driver.wait(
    SeleniumWebdriver.until.elementLocated(locator), timeout);
  await el.click();
};