require("chromedriver");
import chrome from 'selenium-webdriver/chrome';
import SeleniumWebdriver from "selenium-webdriver";

export default function() {
  const options = new chrome.Options().addArguments('headless');

  beforeEach(function() {
    const driver = new SeleniumWebdriver.Builder()
      .forBrowser("chrome")
      .setChromeOptions(
        new chrome.Options().headless())
      .build();
    this.driver = driver;
  });

  afterEach(function() {
    this.driver.quit();
  }); 
}