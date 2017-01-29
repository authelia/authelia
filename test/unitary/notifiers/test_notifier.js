
var sinon = require('sinon');
var Promise = require('bluebird');
var assert = require('assert');

var Notifier = require('../../../src/lib/notifier');
var GmailNotifier = require('../../../src/lib/notifiers/gmail');
var FSNotifier = require('../../../src/lib/notifiers/filesystem');

describe('test notifier', function() {
  it('should build a Gmail Notifier', function() {
    var deps = {};
    deps.nodemailer = {};
    deps.nodemailer.createTransport = sinon.stub().returns({});

    var options = {};
    options.gmail = {};
    options.gmail.user = 'abc';
    options.gmail.pass = 'abcd';

    var notifier = new Notifier(options, deps);
    assert(notifier._notifier instanceof GmailNotifier);
  });

  it('should build a FS Notifier', function() {
    var deps = {};

    var options = {};
    options.filesystem = {};
    options.filesystem.filename = 'abc';

    var notifier = new Notifier(options, deps);
    assert(notifier._notifier instanceof FSNotifier);
  });
});
