const { exec } = require('child_process');
import Bluebird = require("bluebird");

function docker_compose(includes: string[]) {
  const compose_args = includes.map((dc: string) => `-f ${dc}`).join(' ');
  return `docker-compose ${compose_args}`;
}

export class Environment {
  private includes: string[];
  constructor(includes: string[]) {
    this.includes = includes;
  }

  private runCommand(command: string, timeout?: number): Bluebird<void> {
    return new Bluebird<void>((resolve, reject) => {
      console.log('[ENVIRONMENT] Running: %s', command);
      exec(command, (err: any, stdout: any, stderr: any) => {
        if(err) {
          reject(err);
          return;
        }
        if(!timeout) resolve();
        else setTimeout(resolve, timeout);
      });
    });
  }
  

  setup(timeout?: number): Bluebird<void> {
    const command = docker_compose(this.includes) + ' up -d'
    console.log('[ENVIRONMENT] Starting up...');
    return this.runCommand(command, timeout);
  }

  cleanup(): Bluebird<void> {
    if(process.env.KEEP_ENV != "true") {
      const command = docker_compose(this.includes) + ' down'
      console.log('[ENVIRONMENT] Cleaning up...');
      return this.runCommand(command);
    }
    return Bluebird.resolve();
  }

  stop_service(serviceName: string): Bluebird<void> {
    const command = docker_compose(this.includes) + ' stop ' + serviceName;
    console.log('[ENVIRONMENT] Stopping service %s...', serviceName);
    return this.runCommand(command);
  }
  
  start_service(serviceName: string): Bluebird<void> {
    const command = docker_compose(this.includes) + ' start ' + serviceName;
    console.log('[ENVIRONMENT] Starting service %s...', serviceName);
    return this.runCommand(command);
  }
  
  restart_service(serviceName: string, timeout?: number): Bluebird<void> {
    const command = docker_compose(this.includes) + ' restart ' + serviceName;
    console.log('[ENVIRONMENT] Restarting service %s...', serviceName);
    return this.runCommand(command, timeout);
  }
}