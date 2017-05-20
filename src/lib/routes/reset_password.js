
var Promise = require('bluebird');
var objectPath = require('object-path');
var exceptions = require('../Exceptions');
var CHALLENGE = 'reset-password';

var icheck_interface = {
  challenge: CHALLENGE,
  render_template: 'reset-password',
  pre_check_callback: pre_check,
  email_subject: 'Reset your password',
}

module.exports = {
  icheck_interface: icheck_interface,
  post: protect(post)
}

function pre_check(req) {
  var userid = objectPath.get(req, 'body.userid');
  if(!userid) {
    var err = new exceptions.AccessDeniedError();
    return Promise.reject(err);
  }

  var ldap = req.app.get('ldap');

  return ldap.get_emails(userid)
  .then(function(emails) {
    if(!emails && emails.length <= 0) throw new Error('No email found');

    var identity = {}
    identity.email = emails[0];
    identity.userid = userid;
    return Promise.resolve(identity);
  });
}

function protect(fn) {
  return function(req, res) {
    var challenge = objectPath.get(req, 'session.auth_session.identity_check.challenge');
    if(challenge != CHALLENGE) {
      res.status(403);
      res.send();
      return;
    }
    fn(req, res); 
  }
}

function post(req, res) {
  var logger = req.app.get('logger');
  var ldap = req.app.get('ldap');
  var new_password = objectPath.get(req, 'body.password');
  var userid = objectPath.get(req, 'session.auth_session.identity_check.userid');

  logger.info('POST reset-password: User %s wants to reset his/her password', userid);
 
  ldap.update_password(userid, new_password)
  .then(function() {
    logger.info('POST reset-password: Password reset for user %s', userid);
    objectPath.set(req, 'session.auth_session', undefined);
    res.status(204);
    res.send();
  })
  .catch(function(err) {
    logger.error('POST reset-password: Error while resetting the password of user %s. %s', userid, err);
    res.status(500);
    res.send();
  });
}

