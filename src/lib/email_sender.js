
module.exports = EmailSender;

var Promise = require('bluebird');

function EmailSender(nodemailer, options) {
  var transporter = nodemailer.createTransport({
    service: 'gmail',
    auth: {
      user: options.gmail.user,
      pass: options.gmail.pass
    }
  });
  this.transporter = Promise.promisifyAll(transporter);
}

EmailSender.prototype.send = function(to, subject, html) {
  var mailOptions = {};
  mailOptions.from = 'auth-server@open-intent.io';
  mailOptions.to = to;
  mailOptions.subject = subject;
  mailOptions.html = html;
  return this.transporter.sendMailAsync(mailOptions);
}

