
module.exports = first_factor;

var exceptions = require('../exceptions');
var ldap = require('../ldap');
var objectPath = require('object-path');

function replyWithUnauthorized(res) {
    res.status(401);
    res.send();
}

function first_factor(req, res) {
  var logger = req.app.get('logger');
  if(!objectPath.has(req, 'session.auth_session.second_factor')) {
    logger.error('1st factor: Session is missing.');
    replyWithUnauthorized(res);
  }
  
  var username = req.body.username;
  var password = req.body.password;
  console.log('Start authentication of user %s', username);

  if(!username || !password) {
    replyWithUnauthorized(res);
    return;
  }

  logger.info('1st factor: Starting authentication of user "%s"', username);

  var ldap_client = req.app.get('ldap client');
  var config = req.app.get('config');

  logger.debug('1st factor: Start bind operation against LDAP');
  logger.debug('1st factor: username=%s', username);
  logger.debug('1st factor: base_dn=%s', config.ldap_users_dn);
  ldap.validate(ldap_client, username, password, config.ldap_users_dn)
  .then(function() {
    req.session.auth_session.userid = username;
    req.session.auth_session.first_factor = true;
    logger.info('1st factor: LDAP binding successful');
    logger.debug('1st factor: Retrieve email from LDAP');
    return ldap.get_email(ldap_client, username, config.ldap_users_dn)
  })
  .then(function(doc) {
    logger.debug('1st factor: document=%s', JSON.stringify(doc));
    var email = objectPath.get(doc, 'mail');
    logger.debug('1st factor: Retrieved email is %s', email);
    req.session.auth_session.email = email;
    res.status(204);
    res.send();
  })
  .catch(exceptions.LdapSearchError, function(err) {
    logger.info('1st factor: Unable to retrieve email from LDAP');
    res.status(500);
    res.send();
  })
  .catch(exceptions.LdapBindError, function(err) {
    logger.info('1st factor: LDAP binding failed');
    replyWithUnauthorized(res);
  })
  .catch(function(err) {
    logger.debug('1st factor: Unhandled error %s', err);
  });
}
