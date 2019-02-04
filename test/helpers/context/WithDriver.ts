require("chromedriver");
import chrome from 'selenium-webdriver/chrome';
import SeleniumWebdriver from "selenium-webdriver";

export default function() {
  let options = new chrome.Options();

  if (process.env['HEADLESS'] == 'y') {
    options = options.headless();
  }

  beforeEach(function() {
    const driver = new SeleniumWebdriver.Builder()
      .forBrowser("chrome")
      .setChromeOptions(options)
      .build();
    this.driver = driver;
  });

  afterEach(function() {
    this.driver.quit();
  }); 
}