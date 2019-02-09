require("chromedriver");
import chrome from 'selenium-webdriver/chrome';
import SeleniumWebdriver from "selenium-webdriver";

export default function(forEach: boolean = false) {
  let options = new chrome.Options();

  if (process.env['HEADLESS'] == 'y') {
    options = options.headless();
  }

  function beforeBlock(this: Mocha.IHookCallbackContext) {
    const driver = new SeleniumWebdriver.Builder()
      .forBrowser("chrome")
      .setChromeOptions(options)
      .build();
    this.driver = driver;
  }

  function afterBlock(this: Mocha.IHookCallbackContext) {
    return this.driver.quit();
  }

  if (forEach) {
    beforeEach(beforeBlock);
    afterEach(afterBlock);
  } else {
    before(beforeBlock);
    after(afterBlock);
  }
}