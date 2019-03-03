
const userSuite = process.argv[2];

var { setup, teardown } = require(`../test/suites/${userSuite}/environment`);

function sleep(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

let teardownInProgress = false;

process.on('SIGINT', function() {
  if (teardownInProgress) return;
  teardownInProgress = true;
  console.log('Tearing down environment...');
  return teardown()
    .then(() => {
      process.exit(0)
    })
    .catch((err: Error) => {
      console.error(err);
      process.exit(1);
    });
});

function main() {
  console.log('Setting up environment...');
  return setup()
    .then(async () => {
      while (true) {
        await sleep(10000);
      }
    })
    .catch((err: Error) => {
      console.error(err);
      process.exit(1);
    });
}

main();