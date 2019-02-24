
import ChildProcess from 'child_process';
import fs from 'fs';

export default function WithAutheliaRunning(configPath: string, waitTimeout: number = 5000) {
  before(function() {
    this.timeout(10000);

    console.log('Spawning Authelia server with configuration %s.', configPath);
    const authelia = ChildProcess.spawn(
      './scripts/authelia-scripts',
      ['serve', '--no-watch', '--config', configPath],
      {detached: true});

    authelia.on('exit', function(statusCode) {
      console.log('Server terminated with status ' + statusCode);
      if (statusCode != 0) {
        fs.readFile('/tmp/authelia-server.log', function(err, data) {
          if (err) {
            console.error(err);
            return;
          }
          console.log(data.toString('utf-8'));
        })
      }
    });
    this.authelia = authelia;
  
    const waitPromise = new Promise((resolve, reject) => setTimeout(() => resolve(), waitTimeout));
    return waitPromise;
  });
  
  after(function() {
    this.timeout(10000);
  
    console.log('Killing Authelia server.');
    // Kill the group of processes.
    process.kill(-this.authelia.pid);

    // Leave 5 seconds for the process to terminate.
    const waitPromise = new Promise((resolve, reject) => setTimeout(() => resolve(), 5000));
    return waitPromise;
  });
}