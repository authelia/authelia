import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");
import ChildProcess = require("child_process");
import BluebirdPromise = require("bluebird");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  When(/^the application restarts$/, {timeout: 15 * 1000}, function () {
    const exec = BluebirdPromise.promisify(ChildProcess.exec);
    return exec("./scripts/example-commit/dc-example.sh restart authelia && sleep 2");
  });
});