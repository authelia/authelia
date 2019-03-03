import SeleniumWebdriver = require("selenium-webdriver");
import {GetLinkFromFile, GetLinkFromEmail} from './GetIdentityLink';

export default async function(driver: SeleniumWebdriver.WebDriver, email?: boolean, timeout: number = 5000){
  await driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.className("register-totp")), timeout)
  await driver.findElement(SeleniumWebdriver.By.className("register-totp")).click();
  await driver.sleep(500);
  
  const link = (email) ? await GetLinkFromEmail() : await GetLinkFromFile();
  await driver.get(link);
  await driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.className("base32-secret")), timeout);
  return await driver.findElement(SeleniumWebdriver.By.className("base32-secret")).getText();
};
