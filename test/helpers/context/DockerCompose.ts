import { exec } from '../../helpers/utils/exec';

class DockerCompose {
  private commandPrefix: string;

  constructor(composeFiles: string[]) {
    this.commandPrefix = 'docker-compose ' + composeFiles.map((f) => '-f ' + f).join(' ');
  }

  async up() {
    await exec(this.commandPrefix + ' up -d');
  }

  async down() {
    await exec(this.commandPrefix + ' down');
  }

  async restart(service: string) {
    await exec(this.commandPrefix + ' restart ' + service);
  }
}

export default DockerCompose;