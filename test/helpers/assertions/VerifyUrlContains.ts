import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, pattern: string, timeout: number = 5000) {
  await driver.wait(SeleniumWebdriver.until.urlContains(pattern), timeout);
}