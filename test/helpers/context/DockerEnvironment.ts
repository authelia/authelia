import DockerCompose from "./DockerCompose";

class DockerEnvironment {
  private dockerCompose: DockerCompose;

  constructor(composeFiles: string[]) {
    this.dockerCompose = new DockerCompose(composeFiles);
  }

  async start() {
    await this.dockerCompose.up();
  }

  async logs(service: string) {
    await this.dockerCompose.logs(service);
  }

  async stop() {
    await this.dockerCompose.down();
  }
}

export default DockerEnvironment;