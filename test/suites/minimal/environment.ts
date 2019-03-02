import fs from 'fs';
import { exec } from "../../helpers/utils/exec";
import AutheliaServer from "../../helpers/context/AutheliaServer";
import DockerEnvironment from "../../helpers/context/DockerEnvironment";

const autheliaServer = new AutheliaServer(__dirname + '/config.yml');
const dockerEnv = new DockerEnvironment([
  'docker-compose.yml',
  'example/compose/nginx/backend/docker-compose.yml',
  'example/compose/nginx/portal/docker-compose.yml',
  'example/compose/smtp/docker-compose.yml',
])

async function setup() {
  await exec('mkdir -p /tmp/authelia/db');
  await exec('./example/compose/nginx/portal/render.js ' + (fs.existsSync('.suite') ? '': '--production'));
  await dockerEnv.start();
  await autheliaServer.start();
}

async function teardown() {
  await dockerEnv.stop();
  await autheliaServer.stop();
  await exec('mkdir -p /tmp/authelia/db');
}

export { setup, teardown };