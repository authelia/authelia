import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, linkText: string) {
  const element = await driver.wait(
    SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.linkText(linkText)), 5000)
  await element.click();
};