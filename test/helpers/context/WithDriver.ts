require("chromedriver");
import chrome from 'selenium-webdriver/chrome';
import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";

export async function StartDriver() {
  let options = new chrome.Options();

  if (process.env['HEADLESS'] == 'y') {
    options = options.headless();
  }

  const driver = new SeleniumWebdriver.Builder()
    .forBrowser("chrome")
    .setChromeOptions(options)
    .build();
  return driver;
}

export async function StopDriver(driver: WebDriver) {
  return await driver.quit();
}

export default function(forEach: boolean = false) {
  if (forEach) {
    beforeEach(async function() {
      this.driver = await StartDriver();
    });
    afterEach(async function() {
      await StopDriver(this.driver);
    });
  } else {
    before(async function() {
      this.driver = await StartDriver();
    });
    after(async function() {
      await StopDriver(this.driver)
    });
  }
}