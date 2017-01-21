
module.exports = function(user_key_container) {
  return {
    register_request: register_request,
    register: register(user_key_container),
    sign_request: sign_request(user_key_container),
    sign: sign(user_key_container),
  }
}

var objectPath = require('object-path');
var util = require('util');

function replyWithInternalError(res, msg) {
  res.status(500);
  res.send(msg)
}

function replyWithMissingRegistration(res) {
  res.status(401);
  res.send('Please register before authenticate');
}

function replyWithUnauthorized(res) {
  res.status(401);
  res.send();
}


function register_request(req, res) {
  var u2f = req.app.get('u2f');
  var logger = req.app.get('logger');
  var app_id = util.format('https://%s', req.headers.host);

  logger.debug('U2F register_request: headers=%s', JSON.stringify(req.headers));
  logger.info('U2F register_request: Starting registration');
  u2f.startRegistration(app_id, [])
  .then(function(registrationRequest) {
    logger.info('U2F register_request: Sending back registration request');
    req.session.auth_session.register_request = registrationRequest;
    res.status(200);
    res.json(registrationRequest);
  }, function(err) {
    logger.error('U2F register_request: %s', err);
    replyWithInternalError(res, 'Unable to complete the registration');
  });
}

function register(user_key_container) {
  return function(req, res) {
    if(!objectPath.has(req, 'session.auth_session.register_request')) {
      replyWithUnauthorized(res); 
      return; 
    }

    var u2f = req.app.get('u2f');
    var registrationRequest = req.session.auth_session.register_request;
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
      user_key_container[req.session.auth_session.userid] = meta;
      res.status(204);
      res.send();
    }, function(err) {
      logger.error('U2F register: %s', err);
      replyWithInternalError(res, 'Unable to complete the registration');
    });
  }
}

function userKeyExists(req, user_key_container) {
  return req.session.auth_session.userid in user_key_container;
}



function sign_request(user_key_container) {
  return function(req, res) {
    if(!userKeyExists(req, user_key_container)) {
      replyWithMissingRegistration(res);
      return;
    }
  
    var logger = req.app.get('logger');
    var u2f = req.app.get('u2f');
    var key = user_key_container[req.session.auth_session.userid];
    var app_id = util.format('https://%s', req.headers.host);

    logger.info('U2F sign_request: Start authentication');
    u2f.startAuthentication(app_id, [key])
    .then(function(authRequest) {
      logger.info('U2F sign_request: Store authentication request and reply');
      req.session.auth_session.sign_request = authRequest;
      res.status(200);
      res.json(authRequest);
    }, function(err) {
      logger.info('U2F sign_request: %s', err);
      replyWithUnauthorized(res);
    });
  }
}

function sign(user_key_container) {
  return function(req, res) {
    if(!userKeyExists(req, user_key_container)) {
      replyWithMissingRegistration(res);
      return;
    }

    if(!objectPath.has(req, 'session.auth_session.sign_request')) {
      replyWithUnauthorized(res); 
      return; 
    }
  
    var logger = req.app.get('logger');
    var u2f = req.app.get('u2f');
    var authRequest = req.session.auth_session.sign_request;
    var key = user_key_container[req.session.auth_session.userid];
  
    logger.info('U2F sign: Finish authentication');
    u2f.finishAuthentication(authRequest, req.body, [key])
    .then(function(authenticationStatus) {
      logger.info('U2F sign: Authentication successful');
      req.session.auth_session.second_factor = true;
      res.status(204);
      res.send();
    }, function(err) {
      logger.error('U2F sign: %s', err);
      res.status(401);
      res.send();
    });
  }
}
