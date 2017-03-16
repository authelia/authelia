
var Promise = require('bluebird');
var objectPath = require('object-path');
var ldap = require('../ldap');
var exceptions = require('../exceptions');
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

  var ldap_client = req.app.get('ldap client');
  var config = req.app.get('config');

  return ldap.get_email(ldap_client, userid, config.ldap_user_search_base,
                        config.ldap_user_search_filter)
  .then(function(doc) {
    var email = objectPath.get(doc, 'mail');

    var identity = {}
    identity.email = email;
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
  var ldapjs = req.app.get('ldap');
  var ldap_client = req.app.get('ldap client');
  var new_password = objectPath.get(req, 'body.password');
  var userid = objectPath.get(req, 'session.auth_session.identity_check.userid');
  var config = req.app.get('config');

  logger.info('POST reset-password: User %s wants to reset his/her password', userid);
 
  ldap.update_password(ldap_client, ldapjs, userid, new_password, config)
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

