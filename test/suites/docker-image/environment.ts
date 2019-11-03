import DockerEnvironment from "../../helpers/context/DockerEnvironment";
import { exec } from "../../helpers/utils/exec";

const composeFiles = [
  'docker-compose.yml',
  'example/compose/authelia/docker-compose.yml',
  'example/compose/mongo/docker-compose.yml',
  'example/compose/redis/docker-compose.yml',
  'example/compose/nginx/backend/docker-compose.yml',
  'example/compose/nginx/portal/docker-compose.yml',
  'example/compose/smtp/docker-compose.yml',
  'example/compose/httpbin/docker-compose.yml',
  'example/compose/ldap/docker-compose.admin.yml', // This is just used for administration, not for testing.
  'example/compose/ldap/docker-compose.yml'
]

const dockerEnv = new DockerEnvironment(composeFiles);

async function setup() {
  await exec('./example/compose/nginx/portal/render.js --production http://authelia:9091');
  await dockerEnv.start();
}

async function teardown() {
  await dockerEnv.stop();
}

const setup_timeout = 30000;
const teardown_timeout = 30000;

export {
  setup,
  setup_timeout,
  teardown,
  teardown_timeout,
  composeFiles
};