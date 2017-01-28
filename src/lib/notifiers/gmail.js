module.exports = GmailNotifier;

var Promise = require('bluebird');
var fs = require('fs');
var ejs = require('ejs');

var email_template = fs.readFileSync(__dirname + '/../../resources/email-template.ejs', 'UTF-8');

function GmailNotifier(options, deps) {
  var transporter = deps.nodemailer.createTransport({
    service: 'gmail',
    auth: {
      user: options.username,
      pass: options.password
    }
  });
  this.transporter = Promise.promisifyAll(transporter);
}

GmailNotifier.prototype.notify = function(identity, subject, link) {
  var d = {};
  d.url = link;
  d.button_title = 'Continue';
  d.title = subject;

  var mailOptions = {};
  mailOptions.from = 'auth-server@open-intent.io';
  mailOptions.to = identity.email;
  mailOptions.subject = subject;
  mailOptions.html = ejs.render(email_template, d);
  return this.transporter.sendMailAsync(mailOptions);
}

