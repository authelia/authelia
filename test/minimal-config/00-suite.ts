require("chromedriver");
import ChildProcess = require('child_process');
import Bluebird = require("bluebird");

const execAsync = Bluebird.promisify(ChildProcess.exec);

before(function() {
  this.timeout(1000);
  return execAsync("cp users_database.yml users_database.test.yml");
});

after(function() {
  this.timeout(1000);
  return execAsync("rm users_database.test.yml");
});