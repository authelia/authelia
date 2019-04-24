import Chokidar from 'chokidar';
import fs from 'fs';
import ChildProcess from 'child_process';
import kill from 'tree-kill';
import AutheliaServerInterface from './AutheliaServerInterface';
import sleep from '../utils/sleep';

class AutheliaServerWithHotReload implements AutheliaServerInterface {
  private watcher: Chokidar.FSWatcher;
  private configPath: string;
  private AUTHELIA_INTERRUPT_FILENAME = '.authelia-interrupt';
  private serverProcess: ChildProcess.ChildProcess | undefined;
  private clientProcess: ChildProcess.ChildProcess | undefined;
  private filesChangedBuffer: string[] = [];
  private changeInProgress: boolean = false;
  private isInterrupted: boolean = false

  constructor(configPath: string, watchedPaths: string[]) {
    this.configPath = configPath;
    const pathsToReload = ['**/*.go',
      this.AUTHELIA_INTERRUPT_FILENAME, configPath].concat(watchedPaths);
    console.log("Authelia will reload on changes of files or directories in " + pathsToReload.join(', '));
    this.watcher = Chokidar.watch(pathsToReload, {
      persistent: true,
      ignoreInitial: true,
    });
  }

  private async startServer() {
    if (this.serverProcess) return;
    this.serverProcess = ChildProcess.spawn('go',
      ['run', 'main.go', '-config', this.configPath], {
        env: {
          ...process.env,
          NODE_TLS_REJECT_UNAUTHORIZED: "0",
        },
      });
    if (!this.serverProcess || !this.serverProcess.stdout || !this.serverProcess.stderr) return;
    this.serverProcess.stdout.pipe(process.stdout);
    this.serverProcess.stderr.pipe(process.stderr);
    this.serverProcess.on('exit', () => {
      if (!this.serverProcess) return;
      console.log('Authelia server with pid=%s exited.', this.serverProcess.pid);
      this.serverProcess.removeAllListeners();
      this.serverProcess = undefined;
    });
  }

  private killServer() {
    return new Promise((resolve, reject) => {
      if (this.serverProcess) {
        const pid = this.serverProcess.pid;
        try {
          const timeout = setTimeout(
            () => reject(new Error(`Server with pid=${pid} not killed after timeout.`)), 10000);
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
    if (this.clientProcess) return;

    this.clientProcess = ChildProcess.spawn('npm', ['run', 'start'], {
      cwd: './client',
      env: {
        ...process.env,
        'BROWSER': 'none'
      }
    });
    if (!this.clientProcess || !this.clientProcess.stdout || !this.clientProcess.stderr) return;
    this.clientProcess.stdout.pipe(process.stdout);
    this.clientProcess.stderr.pipe(process.stderr);
    this.clientProcess.on('exit', () => {
      if (!this.clientProcess) return;
      console.log('Authelia client exited with pid=%s.', this.clientProcess.pid);
      this.clientProcess.removeAllListeners();
      this.clientProcess = undefined;
    })
  }

  private killClient() {
    return new Promise((resolve, reject) => {
      if (this.clientProcess) {
        const pid = this.clientProcess.pid;
        try {
          const timeout = setTimeout(
            () => reject(new Error(`Server with pid=${pid} not killed after timeout.`)), 10000);
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

  /**
   * Handle file changes.
   * @param path The path of the file that has been changed.
   */
  private onFilesChanged = async (paths: string[]) => {  
    const interruptFileExist = fs.existsSync(this.AUTHELIA_INTERRUPT_FILENAME);
    const interruptFileModified = paths.filter(
      (p) => p === this.AUTHELIA_INTERRUPT_FILENAME).length > 0;
      if (interruptFileExist) {
        if (interruptFileModified) {
          console.log('Authelia is being interrupted.');
          this.isInterrupted = true;
          await this.killServer();
        }
        return;
      } else if (this.isInterrupted && interruptFileModified && !interruptFileExist){
        console.log('Authelia is restarting.');
        await this.startServer();
        this.isInterrupted = false;
        return;
      }

    await this.killServer();
    await this.startServer();

    if (this.filesChangedBuffer.length > 0) {
      await this.consumeFileChanged();
    }
  }

  private async consumeFileChanged() {
    this.changeInProgress = true;
    const paths = this.filesChangedBuffer;
    this.filesChangedBuffer = [];
    try {
      await this.onFilesChanged(paths);
    } catch(e) {
      console.error(e);
    }
    this.changeInProgress = false;
  }

  private enqueueFileChanged(path: string) {
    console.log(`File ${path} has been changed, reloading...`);
    this.filesChangedBuffer.push(path);
    if (this.changeInProgress) return;
    this.consumeFileChanged();
  }

  async start() {
    if (fs.existsSync(this.AUTHELIA_INTERRUPT_FILENAME)) {
      console.log('Authelia is interrupted. Consider removing ' + this.AUTHELIA_INTERRUPT_FILENAME + ' if it\'s not expected.');
      return;
    }

    console.log('Start watching file changes...');
    this.watcher.on('add', (p) => this.enqueueFileChanged(p));
    this.watcher.on('unlink', (p) => this.enqueueFileChanged(p));
    this.watcher.on('change', (p) => this.enqueueFileChanged(p));

    this.startClient();
    this.startServer();
  }

  async stop() {
    await this.killClient();
    await this.killServer();
    await sleep(2000);
  }
}

export default AutheliaServerWithHotReload;