import { exec } from '../../../helpers/utils/exec';

class Kubernetes {
  kubeConfig: string;

  constructor(kubeConfig: string) {
    this.kubeConfig = kubeConfig;
  }

  async apply(configPath: string) {
    await exec('kubectl apply -f ' + configPath, {
      env: {
        KUBECONFIG: this.kubeConfig,
      }
    })
  }

  async loadDockerImage(image: string) {
    await exec('kind load docker-image ' + image), {
      env: {
        KUBECONFIG: this.kubeConfig,
      }
    };
  }
}

export default Kubernetes;