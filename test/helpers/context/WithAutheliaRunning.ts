
import ChildProcess from 'child_process';

export default function WithAutheliaRunning(configPath: string, waitTimeout: number = 3000) {
  before(function() {
    this.timeout(5000);
    const authelia = ChildProcess.spawn(
      './scripts/authelia-scripts',
      ['serve', '--no-watch', '--config', configPath],
      {detached: true});

    authelia.on('exit', function() {
      console.log('Server terminated.');
    });
    this.authelia = authelia;
  
    const waitPromise = new Promise((resolve, reject) => setTimeout(() => resolve(), waitTimeout));
    return waitPromise;
  });
  
  after(function() {
    this.timeout(1000);
  
    // Kill the group of processes.
    process.kill(-this.authelia.pid);
  });
}