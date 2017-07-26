import Cucumber = require("cucumber");

Cucumber.defineSupportCode(function({After}) {
  After(function() {
    return this.driver.quit();
  });
});