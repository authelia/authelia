import SeleniumWebdriver from "selenium-webdriver";

export default async function(driver: any) {
  const content = await driver.wait(
    SeleniumWebdriver.until.elementLocated(
      SeleniumWebdriver.By.tagName('body')), 5000).getText();

  if (content.indexOf('This is a very important secret') > - 1) {
    return;
  }
  else {
    throw new Error('Secret page is not accessible.');
  }
}