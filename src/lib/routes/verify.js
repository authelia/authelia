
module.exports = verify;

var objectPath = require('object-path');
var Promise = require('bluebird');

function verify_filter(req, res) {
  var logger = req.app.get('logger');

  if(!objectPath.has(req, 'session.auth_session'))
    return Promise.reject('No auth_session variable');

  if(!objectPath.has(req, 'session.auth_session.first_factor'))
    return Promise.reject('No first factor variable');

  if(!objectPath.has(req, 'session.auth_session.second_factor'))
    return Promise.reject('No second factor variable');

  if(!objectPath.has(req, 'session.auth_session.userid'))
    return Promise.reject('No userid variable'); 

  var config = req.app.get('config');
  var access_control = config.access_control;

  if(access_control) {
    var allowed_domains = objectPath.get(req, 'session.auth_session.allowed_domains');
    var host = objectPath.get(req, 'headers.host');
    var domain = host.split(':')[0]; 
    logger.debug('Trying to access domain: %s', domain);
    logger.debug('User has access to %s', JSON.stringify(allowed_domains));

    if(allowed_domains.indexOf(domain) < 0)
      return Promise.reject('Access restricted by ACL rules');
  }

  if(!req.session.auth_session.first_factor || 
     !req.session.auth_session.second_factor)
    return Promise.reject('First or second factor not validated');
 
  return Promise.resolve();
}

function verify(req, res) {
  verify_filter(req, res)
  .then(function() {
    res.status(204);
    res.send();
  })
  .catch(function(err) {
    req.app.get('logger').error(err);
    res.status(401);
    res.send();
  });
}

