import Cucumber = require("cucumber");
import fs = require("fs");
import BluebirdPromise = require("bluebird");
import ChildProcess = require("child_process");

Cucumber.defineSupportCode(function({ setDefaultTimeout }) {
  setDefaultTimeout(60 * 1000);
});

Cucumber.defineSupportCode(function({ After, Before }) {
  const exec = BluebirdPromise.promisify(ChildProcess.exec);

  After(function() {
    return this.driver.quit();
  });

  Before({tags: "@needs-test-config", timeout: 15 * 1000}, function () {
    return exec("./scripts/example/dc-example.sh -f docker-compose.test.yml up -d authelia && sleep 2");
  });

  After({tags: "@needs-test-config", timeout: 15 * 1000}, function () {
    return exec("./scripts/example/dc-example.sh up -d authelia && sleep 2");
  });
});