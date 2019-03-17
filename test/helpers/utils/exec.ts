var spawn = require('child_process').spawn;

interface Options {
  cwd?: string;
  env?: {[key: string]: string};
  debug?: boolean;
}

export function exec(command: string, options?: Options): Promise<void> {
  return new Promise((resolve, reject) => {
    const spawnOptions = {
      shell: true,
    } as any;

    if (options && options.cwd) {
      spawnOptions['cwd'] = options.cwd;
    }

    if (options && options.env) {
      spawnOptions['env'] = {
        ...options.env,
        ...process.env,
      }
    }
    
    if (options && options.debug) {
      console.log('>>> ' + command);
    }
    const cmd = spawn(command, spawnOptions);

    cmd.stdout.pipe(process.stdout);
    cmd.stderr.pipe(process.stderr);
    cmd.on('exit', (statusCode: number) => {
      if (statusCode == 0) {
        resolve();
        return;
      }
      reject(new Error('\'' + command + '\' exited with status code ' + statusCode));
    });
  });
}
