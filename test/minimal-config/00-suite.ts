require("chromedriver");
import ChildProcess = require('child_process');
import Bluebird = require("bluebird");

import Environment = require('../environment');

const execAsync = Bluebird.promisify(ChildProcess.exec);

const includes = [
  "docker-compose.minimal.yml",
  "example/compose/docker-compose.base.yml",
  "example/compose/nginx/minimal/docker-compose.yml",
  "example/compose/smtp/docker-compose.yml",
]


before(function() {
  this.timeout(20000);
  this.environment = new Environment.Environment(includes);

  return execAsync("cp users_database.yml users_database.test.yml")
    .then(() => this.environment.setup(2000));
});

after(function() {
  this.timeout(30000);
  return execAsync("rm users_database.test.yml")
    .then(() => {
      if(process.env.KEEP_ENV != "true") {
        return this.environment.cleanup();
      }
    });
});