import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, url: string, timeout: number = 5000) {
  await driver.wait(SeleniumWebdriver.until.urlIs(url), timeout);
}