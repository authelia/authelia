import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, fieldName: string, text: string, timeout: number = 5000) {
  const element = await driver.wait(
    SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.name(fieldName)), timeout)

  await element.sendKeys(text);
};