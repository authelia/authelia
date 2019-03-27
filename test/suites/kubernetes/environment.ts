import KubernetesManager from '../../helpers/context/kubernetes/KubernetesManager';
import Kubernetes from '../../helpers/context/kubernetes/Kubernetes';
import { exec } from '../../helpers/utils/exec';
import WaitUntil from '../../helpers/utils/WaitUntil';
import { spawn, execSync, ChildProcess } from 'child_process';
import treeKill = require('tree-kill');
import Redis, { RedisClient } from 'redis';
import sleep from '../../helpers/utils/sleep';
import DockerEnvironment from '../../helpers/context/DockerEnvironment';

let portFowardingProcess: ChildProcess;

function arePodsReady(kubernetes: Kubernetes): boolean {
  const lines = execSync('kubectl get -n authelia pods --no-headers', { env: {
    KUBECONFIG: kubernetes.kubeConfig,
    ...process.env,
  }}).toString('utf-8').split("\n").filter((x) => x !== '');
  console.log(lines.join('\n'));
  return lines.reduce((acc, line) => {
    return acc && line.indexOf('1/1') > -1;
  }, true);
}

function servicesReady(kubernetes: Kubernetes): Promise<void> {
  return WaitUntil(async () => arePodsReady(kubernetes),
    300000, 15000, 5000);
}

function redisConnected(redisClient: RedisClient): Promise<void> {
  return new Promise<void>((resolve, reject) => {
    console.log('Wait for redis to be connected.');
    redisClient.on('connect', function() {
      resolve();
    });
  })
}

function redisPingOk(redisClient: RedisClient): Promise<boolean> {
  return new Promise<boolean>((resolve, reject) => {
    console.log('Send PING to redis.');
    redisClient.ping((err, msg) => {
      if (err) {
        reject(err);
        return;
      }

      if (msg == 'PONG') {
        resolve(true);
        return;
      }
      resolve(false);
    });
  });
}

async function redisReady(kubernetes: Kubernetes): Promise<void> {
  const redisPortForward = spawn('kubectl',
    ['port-forward', '-n', 'authelia', 'service/redis-service', '8080:6379'], {
    env: {KUBECONFIG: kubernetes.kubeConfig, ...process.env}
  });
  // Wait for the port to be open.
  await sleep(2000);

  const redisClient = Redis.createClient({
    port: 8080,
    no_ready_check: true,
    retry_strategy: () => 3000,
  });

  try {
    await redisConnected(redisClient);
    await WaitUntil(async() => await redisPingOk(redisClient), 30000, 5000);
  } catch(err) {
    console.error(err);
  } finally {
    treeKill(redisPortForward.pid);
  }
}

function startAutheliaPortForwarding(kubernetes: Kubernetes) {
  // Serve applications on port 8080
  portFowardingProcess = spawn('kubectl port-forward --address 0.0.0.0 -n authelia service/nginx-ingress-controller-service 8080:443', {
    shell: true,
    env: {KUBECONFIG: kubernetes.kubeConfig, ...process.env}
  } as any);
  portFowardingProcess.stdout.pipe(process.stdout);
  portFowardingProcess.stderr.pipe(process.stderr);
}

const dockerEnv = new DockerEnvironment([
  'docker-compose.yml',
  'example/compose/nginx/kubernetes/docker-compose.yml',
]);


async function setup() {
  let kubernetes: Kubernetes;
  if (!process.env['KUBECONFIG']) {
    kubernetes = await KubernetesManager.create();
  } else {
    kubernetes = new Kubernetes(process.env['KUBECONFIG'] as string);
  }

  await kubernetes.loadDockerImage('authelia:dist');
  await kubernetes.loadDockerImage('authelia-example-backend');
  
  await exec('./bootstrap.sh', {
    cwd: './example/kube',
    env: {KUBECONFIG: kubernetes.kubeConfig}
  });

  await servicesReady(kubernetes);
  await redisReady(kubernetes);
  await exec('./bootstrap-authelia.sh', {
    cwd: './example/kube',
    env: {KUBECONFIG: kubernetes.kubeConfig}
  });
  await servicesReady(kubernetes);

  await dockerEnv.start();

  startAutheliaPortForwarding(kubernetes);
}

async function teardown() {
  if (portFowardingProcess) {
    console.log('Stopping port forwarding (%s)...', portFowardingProcess.pid);
    treeKill(portFowardingProcess.pid, 'SIGKILL');
    // Wait for the signal to be sent.
    await sleep(1000);
  }

  await dockerEnv.stop();

  if (process.env['KUBECONFIG']) return;
  await KubernetesManager.delete();
}

const setup_timeout = 300000;
const teardown_timeout = 30000;

export {
  setup,
  setup_timeout,
  teardown,
  teardown_timeout
};