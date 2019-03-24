import { exec } from '../../helpers/utils/exec';
import { execSync } from 'child_process';

class DockerCompose {
  private commandPrefix: string;

  constructor(composeFiles: string[]) {
    this.commandPrefix = 'docker-compose ' + composeFiles.map((f) => '-f ' + f).join(' ');
  }

  async up() {
    return await exec(this.commandPrefix + ' up -d');
  }

  async down() {
    return await exec(this.commandPrefix + ' down');
  }

  async restart(service: string) {
    return await exec(this.commandPrefix + ' restart ' + service);
  }

  async ps() {
    return Promise.resolve(execSync(this.commandPrefix + ' ps').toString('utf-8'));
  }

  async logs(service: string) {
    await exec(this.commandPrefix + ' logs ' + service)
  }
}

export default DockerCompose;