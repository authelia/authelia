import AutheliaServerInterface from './AutheliaServerInterface';
import ChildProcess from 'child_process';
import treeKill = require('tree-kill');
import fs from 'fs';

class AutheliaServerFromDist implements AutheliaServerInterface {
  private configPath: string;
  private logInFile: boolean;
  private serverProcess: ChildProcess.ChildProcess | undefined;

  constructor(configPath: string, logInFile: boolean = false) {
    this.configPath = configPath;
    this.logInFile = logInFile;
  }

  async start() {
    console.log("Spawn authelia server from dist.");
    this.serverProcess = ChildProcess.spawn('./scripts/authelia-scripts serve ' + this.configPath, {
      shell: true,
      env: process.env,
    } as any);
    if (!this.serverProcess || !this.serverProcess.stdout || !this.serverProcess.stderr) return;
    if (this.logInFile) {
      var logStream = fs.createWriteStream('/tmp/authelia-server.log', {flags: 'a'});
      this.serverProcess.stdout.pipe(logStream);
      this.serverProcess.stderr.pipe(logStream);
    } else {
      this.serverProcess.stdout.pipe(process.stdout);
      this.serverProcess.stderr.pipe(process.stderr);  
    }
    this.serverProcess.on('exit', (statusCode) => {
      console.log('Authelia server exited with code ' + statusCode);
    })
  }

  async stop() {
    if (!this.serverProcess) return;
    treeKill(this.serverProcess.pid, 'SIGKILL');
  }
}

export default AutheliaServerFromDist;