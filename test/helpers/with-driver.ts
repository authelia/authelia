import SeleniumWebdriver = require("selenium-webdriver");

export default function() {
  before(function() {
    this.driver = new SeleniumWebdriver.Builder()
      .forBrowser("chrome")
      .build();
  })

  after(function() {
    this.driver.quit();
  });
}