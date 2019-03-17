import { exec } from '../../../helpers/utils/exec';
import { execSync } from 'child_process';
import Kubernetes from './Kubernetes';

class KubernetesManager {
  static async create() {
    await exec('kind create cluster');

    const configPath = execSync('kind get kubeconfig-path --name="kind"', {
      env: process.env
    }).toString('utf-8').trim();
    return new Kubernetes(configPath);
  }

  static async delete() {
    await exec('kind delete cluster');
  }
}

export default KubernetesManager;