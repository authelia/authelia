
module.exports = first_factor;

var exceptions = require('../exceptions');
var objectPath = require('object-path');
var Promise = require('bluebird');

function get_allowed_domains(access_control, groups) {
  var allowed_domains = [];

  for(var i = 0; i<access_control.length; ++i) {
    var rule = access_control[i];
    if('group' in rule && 'allowed_domains' in rule) {
      if(groups.indexOf(rule['group']) >= 0) {
        var domains = rule.allowed_domains;
        allowed_domains = allowed_domains.concat(domains);
      }
    }
  }
  return allowed_domains;
}

function first_factor(req, res) {
  var username = req.body.username;
  var password = req.body.password;
  if(!username || !password) {
    res.status(401);
    res.send();
    return;
  }

  var logger = req.app.get('logger');
  var ldap = req.app.get('ldap');
  var config = req.app.get('config');
  var regulator = req.app.get('authentication regulator');

  logger.info('1st factor: Starting authentication of user "%s"', username);
  logger.debug('1st factor: Start bind operation against LDAP');
  logger.debug('1st factor: username=%s', username);

  regulator.regulate(username)
  .then(function() {
    return ldap.bind(username, password);
  })
  .then(function() {
    objectPath.set(req, 'session.auth_session.userid', username);
    objectPath.set(req, 'session.auth_session.first_factor', true);
    logger.info('1st factor: LDAP binding successful');
    logger.debug('1st factor: Retrieve email from LDAP');
    return Promise.join(ldap.get_emails(username), ldap.get_groups(username));
  })
  .then(function(data) {
    var emails = data[0];
    var groups = data[1];

    if(!emails && emails.length <= 0) throw new Error('No email found');
    logger.debug('1st factor: Retrieved email are %s', emails);
    objectPath.set(req, 'session.auth_session.email', emails[0]);

    if(config.access_control) {
      var allowed_domains = get_allowed_domains(config.access_control, groups);
      logger.debug('1st factor: allowed domains are %s', allowed_domains);
      objectPath.set(req, 'session.auth_session.allowed_domains', 
        allowed_domains);
    }
    else {
      logger.debug('1st factor: no access control rules found.' +
        'Default policy to allow all.');
    }

    regulator.mark(username, true);
    res.status(204);
    res.send();
  })
  .catch(exceptions.LdapSearchError, function(err) {
    logger.error('1st factor: Unable to retrieve email from LDAP', err);
    res.status(500);
    res.send();
  })
  .catch(exceptions.LdapBindError, function(err) {
    logger.error('1st factor: LDAP binding failed');
    logger.debug('1st factor: LDAP binding failed due to ', err);
    regulator.mark(username, false);
    res.status(401);
    res.send('Bad credentials');
  })
  .catch(exceptions.AuthenticationRegulationError, function(err) {
    logger.error('1st factor: the regulator rejected the authentication of user %s', username);
    logger.debug('1st factor: authentication rejected due to  %s', err);
    res.status(403);
    res.send('Access has been restricted for a few minutes...');
  })
  .catch(function(err) {
    logger.error('1st factor: Unhandled error %s', err);
    res.status(500);
    res.send('Internal error');
  });
}
