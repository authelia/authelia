
module.exports = Notifier;

var GmailNotifier = require('./notifiers/gmail.js');
var FSNotifier = require('./notifiers/filesystem.js');

function notifier_factory(options, deps) {
  if('gmail' in options) {
    return new GmailNotifier(options.gmail, deps);
  }
  else if('filesystem' in options) {
    return new FSNotifier(options.filesystem);
  }
}

function Notifier(options, deps) {
  this._notifier = notifier_factory(options, deps);
}

Notifier.prototype.notify = function(identity, subject, link) {
  return this._notifier.notify(identity, subject, link); 
}


