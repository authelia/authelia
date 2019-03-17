import { exec } from './utils/exec';

const userSuite = process.argv[2];
const command = process.argv[3]; // The command to run once the env is up.

var { setup, setup_timeout, teardown, teardown_timeout } = require(`../test/suites/${userSuite}/environment`);

function sleep(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

let teardownInProgress = false;

async function block() {
  while (true) {
    await sleep(10000);
  }
}

async function blockOrRun(cmd: string | null) {
  if (cmd) {
    await exec(cmd);
  } else {
    await block();
  }
}

process.on('SIGINT', function() {
  if (teardownInProgress) return;
  teardownInProgress = true;

  stop()
    .then(() => process.exit(0))
    .catch(() => process.exit(1));
});

async function stop() {
  const timer = setTimeout(() => {
    console.error('Teardown timed out...');
    process.exit(1);
  }, teardown_timeout);
  console.log('>>> Tearing down environment <<<');
  try {  
    await teardown();
    clearTimeout(timer);
  } catch (err) {
    console.error(err);
    throw err;
  }
}

async function start() {
  const timer = setTimeout(() => {
    console.error('Setup timed out...');
    teardown().finally(() => process.exit(1));
  }, setup_timeout);
  console.log('>>> Setting up environment <<<');
  try {
    await setup();
    clearTimeout(timer);
    await blockOrRun(command);
    if (!teardownInProgress) {
      await stop();
      process.exit(0);
    }
  }
  catch (err) {
    console.error(err);
    await stop();
    process.exit(1);
  }
}

start();