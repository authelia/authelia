import SeleniumWebdriver, { WebDriver, Locator } from "selenium-webdriver";

export default async function(driver: WebDriver, locator: Locator) {
  const el = await driver.wait(
    SeleniumWebdriver.until.elementLocated(locator), 5000);

  await el.click();
};