require("chromedriver");
import Environment = require('../environment');

const includes = [
  "docker-compose.yml",
  "example/compose/docker-compose.base.yml",
  "example/compose/mongo/docker-compose.yml",
  "example/compose/redis/docker-compose.yml",
  "example/compose/nginx/backend/docker-compose.yml",
  "example/compose/nginx/portal/docker-compose.yml",
  "example/compose/smtp/docker-compose.yml",
  "example/compose/httpbin/docker-compose.yml",
  "example/compose/ldap/docker-compose.yml"
];


before(function() {
  this.timeout(20000);
  this.environment = new Environment.Environment(includes);
  return this.environment.setup(5000);
});

after(function() {
  this.timeout(30000);
  if(process.env.KEEP_ENV != "true") {
    return this.environment.cleanup();
  }
});