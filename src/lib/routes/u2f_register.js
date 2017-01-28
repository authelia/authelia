
var u2f_register_handler = require('./u2f_register_handler');

module.exports = {
  register_request: register_request,
  register: register,
  register_handler_get: u2f_register_handler.get,
  register_handler_post: u2f_register_handler.post
}

var objectPath = require('object-path');
var u2f_common = require('./u2f_common');
var Promise = require('bluebird');

function register_request(req, res) {
  var logger = req.app.get('logger');
  var challenge = objectPath.get(req, 'session.auth_session.identity_check.challenge');
  if(challenge != 'u2f-register') {
    res.status(403);
    res.send();
    return;
  }

  var u2f = req.app.get('u2f');
  var appid = u2f_common.extract_app_id(req);

  logger.debug('U2F register_request: headers=%s', JSON.stringify(req.headers));
  logger.info('U2F register_request: Starting registration');
  u2f.startRegistration(appid, [])
  .then(function(registrationRequest) {
    logger.info('U2F register_request: Sending back registration request');
    req.session.auth_session.register_request = registrationRequest;
    res.status(200);
    res.json(registrationRequest);
  })
  .catch(function(err) {
    logger.error('U2F register_request: %s', err);
    res.status(500);
    res.send('Unable to start registration request');
  });
}

function register(req, res) {
  var registrationRequest = objectPath.get(req, 'session.auth_session.register_request');
  var challenge = objectPath.get(req, 'session.auth_session.identity_check.challenge');

  if(!(registrationRequest && challenge == 'u2f-register')) {
    res.status(403);
    res.send();
    return; 
  }

  var user_data_storage = req.app.get('user data store');
  var u2f = req.app.get('u2f');
  var userid = req.session.auth_session.userid;
  var appid = u2f_common.extract_app_id(req);
  var logger = req.app.get('logger');

  logger.info('U2F register: Finishing registration');
  logger.debug('U2F register: register_request=%s', JSON.stringify(registrationRequest));
  logger.debug('U2F register: body=%s', JSON.stringify(req.body));

  u2f.finishRegistration(registrationRequest, req.body)
  .then(function(registrationStatus) {
    logger.info('U2F register: Store registration and reply');
    var meta = {
      keyHandle: registrationStatus.keyHandle,
      publicKey: registrationStatus.publicKey,
      certificate: registrationStatus.certificate
    }
    return user_data_storage.set_u2f_meta(userid, appid, meta);
  })
  .then(function() {
    objectPath.set(req, 'session.auth_session.identity_check', undefined);
    res.status(204);
    res.send();
  })
  .catch(function(err) {
    logger.error('U2F register: %s', err);
    res.status(500);
    res.send('Unable to register');
  });
}

