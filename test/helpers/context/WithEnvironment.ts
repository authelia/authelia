import sleep from '../utils/sleep';

export default function WithEnvironment(suite: string, waitTimeout: number = 5000) {
  var { setup, teardown } = require(`../../suites/${suite}/environment`);

  before(async function() {
    this.timeout(30000);

    console.log('Preparing environment...');
    await setup();  
    await sleep(waitTimeout);
  });
  
  after(async function() {
    this.timeout(30000);

    console.log('Stopping environment...');
    await teardown();

    await sleep(waitTimeout);
  });
}