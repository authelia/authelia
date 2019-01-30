import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

export default async function(
  driver: WebDriver,
  username: string,
  password: string,
  keepMeLoggedIn: boolean = false) {
  
  await driver.wait(SeleniumWebdriver.until.elementLocated(SeleniumWebdriver.By.id("username")), 5000)
  await driver.findElement(SeleniumWebdriver.By.id("username")).sendKeys(username);
  await driver.findElement(SeleniumWebdriver.By.id("password")).sendKeys(password);
  if (keepMeLoggedIn) {
    await driver.findElement(SeleniumWebdriver.By.id("keep_me_logged_in")).click();
    return;
  }
  await driver.findElement(SeleniumWebdriver.By.tagName("button")).click();
};