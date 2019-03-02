import sleep from '../utils/sleep';

export default function WithAutheliaRunning(suitePath: string, waitTimeout: number = 5000) {
  const suite = suitePath.split('/').slice(-1)[0];
  var { setup, teardown } = require(`../../suites/${suite}/environment`);

  before(async function() {
    this.timeout(10000);

    console.log('Preparing environment...');
    await setup();  
    await sleep(waitTimeout);
  });
  
  after(async function() {
    this.timeout(10000);

    console.log('Stopping environment...');
    await teardown();

    await sleep(waitTimeout);
  });
}