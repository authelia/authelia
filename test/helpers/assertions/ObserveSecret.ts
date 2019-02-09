import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

// Verify if the current page contains "This is a very important secret!".
export default async function(driver: WebDriver, timeout: number = 5000) {
  const el = await driver.wait(
    SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.tagName('body')), timeout);

  await driver.wait(
    SeleniumWebdriver.until.elementTextContains(el, "This is a very important secret!"), timeout);
}