var sinon = require('sinon');
var assert = require('assert');
var GmailNotifier = require('../../../src/lib/notifiers/gmail');

describe('test gmail notifier', function() {
  it('should send an email', function() {
    var nodemailer = {};
    var transporter = {};
    nodemailer.createTransport = sinon.stub().returns(transporter);
    transporter.sendMail = sinon.stub().yields();
    var options = {};
    options.username = 'user_gmail';
    options.password = 'pass_gmail';

    var deps = {};
    deps.nodemailer = nodemailer;
    
    var sender = new GmailNotifier(options, deps);
    var subject = 'subject';

    var identity = {};
    identity.userid = 'user';
    identity.email = 'user@example.com';

    var url = 'http://test.com';

    return sender.notify(identity, subject, url)
    .then(function() {
      assert.equal(nodemailer.createTransport.getCall(0).args[0].auth.user, 'user_gmail');
      assert.equal(nodemailer.createTransport.getCall(0).args[0].auth.pass, 'pass_gmail');
      assert.equal(transporter.sendMail.getCall(0).args[0].to, 'user@example.com');
      assert.equal(transporter.sendMail.getCall(0).args[0].subject, 'subject');
      return Promise.resolve();
    });
  });
});
