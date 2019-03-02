import ChildProcess from 'child_process';
import Bluebird from 'bluebird';
import Assert from 'assert';
import sleep from '../../helpers/utils/sleep';
import AutheliaSuite from '../../helpers/context/AutheliaSuite';

const execAsync = Bluebird.promisify<string, string>(ChildProcess.exec);

AutheliaSuite('Test docker container runs as expected', __dirname, function() {
  this.timeout(15000);
  
  it('should be running', async function() {
    await sleep(5000);
    const output: string = await execAsync('docker ps -a | grep "authelia-test"');
    Assert(output.match(new RegExp('Up [0-9]+ seconds')));
  });
});