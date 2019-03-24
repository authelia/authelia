import SeleniumWebDriver, { WebDriver } from "selenium-webdriver";

/**
 * 
 * @param driver The selenium web driver.
 * @param locator The locator of the element to find in the DOM.
 */
export default async function(driver: WebDriver, locator: SeleniumWebDriver.Locator) {
  const els = await driver.findElements(locator);
  if (els.length == 0) {
    throw new Error("Element does not exist.");
  }
}