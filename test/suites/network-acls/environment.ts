import fs from 'fs';
import { exec } from "../../helpers/utils/exec";
import AutheliaServer from "../../helpers/context/AutheliaServer";
import DockerEnvironment from "../../helpers/context/DockerEnvironment";

const autheliaServer = new AutheliaServer(__dirname + '/config.yml', [__dirname + '/users_database.yml']);
const dockerEnv = new DockerEnvironment([
  'docker-compose.yml',
  'example/compose/nginx/backend/docker-compose.yml',
  'example/compose/nginx/portal/docker-compose.yml',
  'example/compose/squid/docker-compose.yml',
  'example/compose/smtp/docker-compose.yml',
  // To debug headers
  'example/compose/httpbin/docker-compose.yml',
])

async function setup() {
  await exec(`cp ${__dirname}/users_database.yml ${__dirname}/users_database.test.yml`);
  await exec('./example/compose/nginx/portal/render.js ' + (fs.existsSync('.suite') ? '': '--production'));
  await dockerEnv.start();
  await autheliaServer.start();
}

async function teardown() {
  await autheliaServer.stop();
  await dockerEnv.stop();
}

const setup_timeout = 30000;
const teardown_timeout = 30000;

export {
  setup,
  setup_timeout,
  teardown,
  teardown_timeout
};