
module.exports = first_factor;

var exceptions = require('../exceptions');
var ldap = require('../ldap');
var objectPath = require('object-path');

function first_factor(req, res) {
  var logger = req.app.get('logger');
  var username = req.body.username;
  var password = req.body.password;
  if(!username || !password) {
    res.status(401);
    res.send();
    return;
  }

  logger.info('1st factor: Starting authentication of user "%s"', username);

  var ldap_client = req.app.get('ldap client');
  var config = req.app.get('config');
  var regulator = req.app.get('authentication regulator');

  logger.debug('1st factor: Start bind operation against LDAP');
  logger.debug('1st factor: username=%s', username);
  logger.debug('1st factor: base_dn=%s', config.ldap_users_dn);

  regulator.regulate(username)
  .then(function() {
    return ldap.validate(ldap_client, username, password, config.ldap_users_dn);
  })
  .then(function() {
    objectPath.set(req, 'session.auth_session.userid', username);
    objectPath.set(req, 'session.auth_session.first_factor', true);
    logger.info('1st factor: LDAP binding successful');
    logger.debug('1st factor: Retrieve email from LDAP');
    return ldap.get_email(ldap_client, username, config.ldap_users_dn)
  })
  .then(function(doc) {
    var email = objectPath.get(doc, 'mail');
    logger.debug('1st factor: document=%s', JSON.stringify(doc));
    logger.debug('1st factor: Retrieved email is %s', email);

    objectPath.set(req, 'session.auth_session.email', email);
    regulator.mark(username, true);
    res.status(204);
    res.send();
  })
  .catch(exceptions.LdapSearchError, function(err) {
    logger.info('1st factor: Unable to retrieve email from LDAP', err);
    res.status(500);
    res.send();
  })
  .catch(exceptions.LdapBindError, function(err) {
    logger.info('1st factor: LDAP binding failed');
    logger.debug('1st factor: LDAP binding failed due to ', err);
    regulator.mark(username, false);
    res.status(401);
    res.send('Bad credentials');
  })
  .catch(exceptions.AuthenticationRegulationError, function(err) {
    logger.info('1st factor: the regulator rejected the authentication of user %s', username);
    logger.debug('1st factor: authentication rejected due to  %s', err);
    res.status(403);
    res.send('Access has been restricted for a few minutes...');
  })
  .catch(function(err) {
    logger.debug('1st factor: Unhandled error %s', err);
    res.status(500);
    res.send('Internal error');
  });
}
