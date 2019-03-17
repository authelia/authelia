import { exec } from 'child_process';

function execPromise(command: string) {
  return new Promise<string>(function(resolve, reject) {
      exec(command, (error, stdout, stderr) => {
          if (error) {
              reject(error);
              return;
          }

          resolve(stdout.trim());
      });
  });
}

export default execPromise;