import fs from 'fs';
import AutheliaServerWithHotReload from "./AutheliaServerWithHotReload";
import AutheliaServerInterface from './AutheliaServerInterface';
import AutheliaServerFromDist from './AutheliaServerFromDist';

class AutheliaServer implements AutheliaServerInterface {
  private runnerImpl: AutheliaServerInterface;

  constructor(configPath: string, watchPaths: string[] = []) {
    if (fs.existsSync('.suite')) {
      this.runnerImpl = new AutheliaServerWithHotReload(configPath, watchPaths);
    } else {
      this.runnerImpl = new AutheliaServerFromDist(configPath, true);
    }
  }

  async start() {
    await this.runnerImpl.start();
  }

  async stop() {
    await this.runnerImpl.stop();
  }
}

export default AutheliaServer;