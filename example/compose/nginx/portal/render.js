#!/usr/bin/env node

const ejs = require('ejs');
const fs = require('fs');
const program = require('commander');

let backend;

program
  .version('0.1.0')
  .option('-p, --production', 'Render template for production.')
  .arguments('[backend]')
  .action((backendArg) => backend = backendArg)
  .parse(process.argv)

const options = {
  production: false,
}

if (!backend) {
  backend = 'http://192.168.240.1:9091'
}

if (program.production) {
  options['production'] = true;
}

options['authelia_backend'] = backend;

const templatePath = __dirname + '/nginx.conf.ejs';
const outputPath = __dirname + '/nginx.conf';

html = ejs.renderFile(templatePath, options, (err, conf) => {
  try {
    var fd = fs.openSync(outputPath, 'w');
    fs.writeFileSync(fd, conf);
  } catch (e) {
    fs.writeFileSync(outputPath, conf);
  }
});