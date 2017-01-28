module.exports = FSNotifier;

var Promise = require('bluebird');
var fs = Promise.promisifyAll(require('fs'));
var util = require('util');

function FSNotifier(options) {
  this._filename = options.filename;
}

FSNotifier.prototype.notify = function(identity, subject, link) {
  var content = util.format('User: %s\nSubject: %s\nLink: %s', identity.userid,
    subject, link); 
  return fs.writeFileAsync(this._filename, content);
}

