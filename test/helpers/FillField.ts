import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, fieldName: string, text: string) {
  const element = await driver.wait(
    SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.name(fieldName)), 5000)

  await element.sendKeys(text);
};