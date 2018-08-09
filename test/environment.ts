const { exec } = require('child_process');
import Bluebird = require("bluebird");

function docker_compose(includes: string[]) {
  const compose_args = includes.map((dc: string) => `-f ${dc}`).join(' ');
  return `docker-compose ${compose_args}`;
}

export function setup(includes: string[], setupTime: number = 2000): Bluebird<void> {
  const command = docker_compose(includes) + ' up -d'
  console.log('Starting up environment.');
  console.log('Running: %s', command);

  return new Bluebird<void>(function(resolve, reject) {
      exec(command, function(err, stdout, stderr) {
      if(err) {
        reject(err);
        return;
      }
      setTimeout(function() {
        resolve();
      }, setupTime);
    });
  });
}

export function cleanup(includes: string[]): Bluebird<void> {
  const command = docker_compose(includes) + ' down'; 
  console.log('Shutting down environment.');
  console.log('Running: %s', command);

  return new Bluebird<void>(function(resolve, reject) {
    exec(command, function(err, stdout, stderr) {
      if(err) {
        reject(err);
        return;
      }
      resolve();
    });
  });
}