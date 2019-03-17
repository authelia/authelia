import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, match: string, timeout: number = 5000) {
  const el = await driver.wait(
    SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.tagName('body')), timeout);

  await driver.wait(
    SeleniumWebdriver.until.elementTextContains(el, match), timeout);
}