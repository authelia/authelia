require("chromedriver");
import Environment = require('../environment');

const includes = [
  "docker-compose.minimal.yml",
  "example/compose/docker-compose.base.yml",
  "example/compose/nginx/minimal/docker-compose.yml",
  "example/compose/ldap/docker-compose.yml"
]


before(function() {
  this.timeout(20000);
  return Environment.setup(includes);
});

after(function() {
  this.timeout(30000);
  return Environment.cleanup(includes);
});