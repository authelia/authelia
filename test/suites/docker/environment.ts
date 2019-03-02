import { exec } from '../../helpers/utils/exec';
import ChildProcess from 'child_process';

async function setup() {
  await exec('docker run -d -v $(pwd)/config.yml:/etc/authelia/config.yml --name authelia-test clems4ever/authelia > /dev/null');
  console.log('Container has been spawned.');
}

async function teardown() {
  try {
    ChildProcess.execSync('docker ps | grep "authelia-test"');
    await exec('docker rm -f authelia-test > /dev/null');
  } catch (e) {
    // If grep does not find anything, execSync throws an exception since the command returns 1.
  }
}

export { setup, teardown };