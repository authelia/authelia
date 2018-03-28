import {When} from "cucumber";
import seleniumWebdriver = require("selenium-webdriver");

When("I sleep for {int} seconds", function (seconds: number) {
  return this.driver.sleep(seconds * 1000);
});
