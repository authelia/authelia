#!/usr/bin/env node

const ejs = require('ejs');
const fs = require('fs');
const program = require('commander');

program
  .version('0.1.0')
  .option('-p, --production', 'Render template for production.')
  .parse(process.argv)

const options = {
  production: false,
}

if (program.production) {
  options['production'] = true;
}

html = ejs.renderFile(__dirname + '/nginx.conf.ejs', options, (err, conf) => {
  fs.writeFileSync(__dirname + '/nginx.conf', conf);
});