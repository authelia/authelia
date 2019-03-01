import ChildProcess from 'child_process';
import Bluebird from 'bluebird';
import Assert from 'assert';

const execAsync = Bluebird.promisify<string, string>(ChildProcess.exec);

function sleep(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

describe('Test docker container can run', function() {
  this.timeout(15000);

  before(async function() {
    await execAsync('docker run -d -v $(pwd)/config.yml:/etc/authelia/config.yml --name authelia-test clems4ever/authelia');
  });
  
  after(async function() {
    await execAsync('docker rm -f authelia-test');
  });
  
  it('should be running', async function() {
    await sleep(5000);
    const output: string = await execAsync('docker ps -a | grep "authelia-test"');
    Assert(output.match(new RegExp('Up [0-9] seconds')));
  });
});