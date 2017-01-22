

var sinon = require('sinon');
var assert = require('assert');
var EmailSender = require('../../src/lib/email_sender');

describe('test email sender', function() {
  it('should send an email', function() {
    var nodemailer = {};
    var transporter = {};
    nodemailer.createTransport = sinon.stub().returns(transporter);
    transporter.sendMail = sinon.stub().yields();
    var options = {};
    options.gmail = {};
    options.gmail.user = 'test@gmail.com';
    options.gmail.pass = 'test@gmail.com';
    
    var sender = new EmailSender(nodemailer, options);
    var to = 'example@gmail.com';
    var subject = 'subject';
    var content = 'content';

    return sender.send(to, subject, content)
    .then(function() {
      assert.equal(to, transporter.sendMail.getCall(0).args[0].to);
      assert.equal(subject, transporter.sendMail.getCall(0).args[0].subject);
      assert.equal(content, transporter.sendMail.getCall(0).args[0].html);
      return Promise.resolve();
    });
  });
});
