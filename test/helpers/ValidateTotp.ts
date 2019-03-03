import Speakeasy from "speakeasy";
import SeleniumWebdriver, { WebDriver } from 'selenium-webdriver';

export default async function(driver: WebDriver, secret: string, timeout: number = 5000) {
  const token = Speakeasy.totp({
    secret: secret,
    encoding: "base32"
  });

  await driver.wait(SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.id("totp-token")), timeout)
  await driver.findElement(SeleniumWebdriver.By.id("totp-token")).sendKeys(token);
  
  const el = await driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.id('totp-button')), timeout);
  el.click();
}