require("chromedriver");
import Bluebird = require("bluebird");
import Configuration = require("../configuration");
import Environment = require("../environment");

import ChildProcess = require('child_process');
const execAsync = Bluebird.promisify(ChildProcess.exec);

const includes = [
  "docker-compose.test.yml",
  "example/compose/docker-compose.base.yml",
  "example/compose/nginx/minimal/docker-compose.yml",
  "example/compose/smtp/docker-compose.yml",
]


before(function() {
  this.timeout(20000);
  this.environment = new Environment.Environment(includes);
  this.configuration = new Configuration.Configuration();

  return this.configuration.setup(
    "config.minimal.yml",
    "config.test.yml",
    conf => {
      conf.session.inactivity = 2000;
    })
    .then(() => execAsync("cp users_database.yml users_database.test.yml"))
    .then(() => this.environment.setup(2000));
});

after(function() {
  this.timeout(30000);
  return this.configuration.cleanup()
    .then(() => execAsync("rm users_database.test.yml"))
    .then(() => this.environment.cleanup());
});