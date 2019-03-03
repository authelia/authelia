import Chokidar from 'chokidar';
import fs from 'fs';
import { exec } from '../utils/exec';
import ChildProcess from 'child_process';
import kill from 'tree-kill';
import AutheliaServerInterface from './AutheliaServerInterface';

class AutheliaServerWithHotReload implements AutheliaServerInterface {
  private watcher: Chokidar.FSWatcher;
  private configPath: string;
  private AUTHELIA_INTERRUPT_FILENAME = '.authelia-interrupt';
  private serverProcess: ChildProcess.ChildProcess | undefined;
  private clientProcess: ChildProcess.ChildProcess | undefined;

  constructor(configPath: string) {
    this.configPath = configPath;
    this.watcher = Chokidar.watch(['server', 'shared/**/*.ts', 'node_modules',
      this.AUTHELIA_INTERRUPT_FILENAME, configPath], {
      persistent: true,
      ignoreInitial: true,
    });
  }

  private async startServer() {
    await exec('./node_modules/.bin/tslint -c server/tslint.json -p server/tsconfig.json')
    this.serverProcess = ChildProcess.spawn('./node_modules/.bin/ts-node',
      ['-P', './server/tsconfig.json', './server/src/index.ts', this.configPath], {
        env: {...process.env},
      });
    this.serverProcess.stdout.pipe(process.stdout);
    this.serverProcess.stderr.pipe(process.stderr);
    this.serverProcess.on('exit', () => {
      console.log('Authelia server exited.');
      if (!this.serverProcess) return;
      this.serverProcess.removeAllListeners();
      this.serverProcess = undefined;
    });
  }

  private killServer() {
    return new Promise((resolve, reject) => {
      if (this.serverProcess) {
        try {
          const timeout = setTimeout(
            () => reject(new Error('Server not killed after timeout.')), 10000);
          this.serverProcess.on('exit', () => {
            clearTimeout(timeout);
            resolve();
          });
          kill(this.serverProcess.pid, 'SIGKILL');
        } catch (e) {
          reject(e);
        }
      } else {
        resolve();
      }
    });
  }

  private async startClient() {
    this.clientProcess = ChildProcess.spawn('npm', ['run', 'start'], {
      cwd: './client',
      env: {
        ...process.env,
        'BROWSER': 'none'
      }
    });
    this.clientProcess.stdout.pipe(process.stdout);
    this.clientProcess.stderr.pipe(process.stderr);
    this.clientProcess.on('exit', () => {
      console.log('Authelia client exited.');
      if (!this.clientProcess) return;
      this.clientProcess.removeAllListeners();
      this.clientProcess = undefined;
    })
  }

  private killClient() {
    return new Promise((resolve, reject) => {
      if (this.clientProcess) {
        try {
          const timeout = setTimeout(
            () => reject(new Error('Server not killed after timeout.')), 10000);
          this.clientProcess.on('exit', () => {
            clearTimeout(timeout);
            resolve();
          });
          kill(this.clientProcess.pid, 'SIGKILL');
        } catch (e) {
          reject(e);
        }
      } else {
        resolve();
      }
    });
  }

  private async generateConfigurationSchema() {
    await exec('./node_modules/.bin/typescript-json-schema -o ' +
                'server/src/lib/configuration/Configuration.schema.json ' +
                '--strictNullChecks --required server/tsconfig.json Configuration');
  }

  /**
   * Handle file changes.
   * @param path The path of the file that has been changed.
   */
  private async onFileChanged(path: string) {
    console.log(`File ${path} has been changed, reloading...`);
    if (path.startsWith('server/src/lib/configuration/schema')) {
      console.log('Schema needs to be regenerated.');
      await this.generateConfigurationSchema();
    }
    else if (path === this.AUTHELIA_INTERRUPT_FILENAME) {
      if (fs.existsSync(path)) {
        console.log('Authelia is being interrupted.');
        await this.killServer();
      } else {
        console.log('Authelia is restarting.');
        await this.startServer();
      }
      return;
    }
    await this.killServer();
    await this.startServer();
  }

  async start() {
    if (fs.existsSync(this.AUTHELIA_INTERRUPT_FILENAME)) {
      console.log('Authelia is interrupted. Consider removing ' + this.AUTHELIA_INTERRUPT_FILENAME + ' if it\'s not expected.');
      return;
    }

    console.log('Start watching file changes...');
    this.watcher.on('add', (p) => this.onFileChanged(p));
    this.watcher.on('unlink', (p) => this.onFileChanged(p));
    this.watcher.on('change', (p) => this.onFileChanged(p));

    this.startClient();
    this.startServer();
  }

  async stop() {
    await this.killClient();
    await this.killServer();
  }
}

export default AutheliaServerWithHotReload;