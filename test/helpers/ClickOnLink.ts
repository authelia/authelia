import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, linkText: string, timeout: number = 5000) {
  const element = await driver.wait(
    SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.linkText(linkText)), timeout);
  await element.click();
};