var spawn = require('child_process').spawn;

function exec(cmd) {
  return new Promise((resolve, reject) => {
    const command = spawn(cmd, {
      shell: true
    });
    command.stdout.pipe(process.stdout);
    command.stderr.pipe(process.stderr);

    command.on('exit', function(statusCode) {
      if (statusCode != 0) {
        reject(new Error('Exited with status ' + statusCode));
        return;
      }
      resolve();
    })
  })
}

module.exports = { exec }