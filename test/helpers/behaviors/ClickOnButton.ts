import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, text: string, timeout: number = 5000) {
  const element = await driver.wait(
    SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.xpath("//button[text()='" + text + "']")), timeout)
  await element.click();
};