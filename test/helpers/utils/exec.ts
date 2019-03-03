var spawn = require('child_process').spawn;


export function exec(command: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const cmd = spawn(command, {
      shell: true,
    });

    cmd.stdout.pipe(process.stdout);
    cmd.stderr.pipe(process.stderr);
    cmd.on('exit', (statusCode: number) => {
      if (statusCode == 0) {
        resolve();
        return;
      }
      reject(new Error('Exited with status code ' + statusCode));
    });
  });
}
