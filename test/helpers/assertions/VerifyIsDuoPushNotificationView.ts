import SeleniumWebDriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, timeout: number = 5000) {
  await driver.wait(SeleniumWebDriver.until.elementLocated(
    SeleniumWebDriver.By.className('duo-push-view')), timeout);
}