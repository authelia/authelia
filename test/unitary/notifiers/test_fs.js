var sinon = require('sinon');
var assert = require('assert');
var FSNotifier = require('../../../src/lib/notifiers/filesystem');
var tmp = require('tmp');
var fs = require('fs');

describe('test FS notifier', function() {
  var tmpDir;
  before(function() {
    tmpDir = tmp.dirSync({ unsafeCleanup: true });
  });

  after(function() {
    tmpDir.removeCallback();
  });

  it('should write the notification in a file', function() {
    var options = {};
    options.filename = tmpDir.name + '/notification';

    var sender = new FSNotifier(options);
    var subject = 'subject';

    var identity = {};
    identity.userid = 'user';
    identity.email = 'user@example.com';

    var url = 'http://test.com';

    return sender.notify(identity, subject, url)
    .then(function() {
      var content = fs.readFileSync(options.filename, 'UTF-8');
      assert(content.length > 0);
      return Promise.resolve();
    });
  });
});
