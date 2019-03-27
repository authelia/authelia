require("chromedriver");
import chrome from 'selenium-webdriver/chrome';
import SeleniumWebdriver, { WebDriver, ProxyConfig } from "selenium-webdriver";

export async function StartDriver(proxy?: ProxyConfig) {
  let options = new chrome.Options();

  if (process.env['HEADLESS'] == 'y') {
    options = options.headless();
  }

  let driverBuilder = new SeleniumWebdriver.Builder()
    .forBrowser("chrome");

  if (proxy) {
    options = options.addArguments(`--proxy-server=${proxy.httpProxy}`)
  }

  driverBuilder = driverBuilder.setChromeOptions(options);
  return await driverBuilder.build();
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