
module.exports = {
  register_request: register_request,
  register: register,
  sign_request: sign_request,
  sign: sign,
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

function extractAppId(req) {
  return util.format('https://%s', req.headers.host);
}

function register_request(req, res) {
  var u2f = req.app.get('u2f');
  var logger = req.app.get('logger');
  var appid = extractAppId(req);

  logger.debug('U2F register_request: headers=%s', JSON.stringify(req.headers));
  logger.info('U2F register_request: Starting registration');
  u2f.startRegistration(appid, [])
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

function register(req, res) {
  if(!objectPath.has(req, 'session.auth_session.register_request')) {
    replyWithUnauthorized(res); 
    return; 
  }

  var user_data_storage = req.app.get('user data store');
  var u2f = req.app.get('u2f');
  var registrationRequest = req.session.auth_session.register_request;
  var userid = req.session.auth_session.userid;
  var appid = extractAppId(req);
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
    user_data_storage.set_u2f_meta(userid, appid, meta);
    res.status(204);
    res.send();
  }, function(err) {
    logger.error('U2F register: %s', err);
    replyWithInternalError(res, 'Unable to complete the registration');
  });
}

function retrieveU2fMeta(req, user_data_storage) {
  var userid = req.session.auth_session.userid;
  var appid = extractAppId(req);
  return user_data_storage.get_u2f_meta(userid, appid);
}

function startU2fAuthentication(u2f, appid, meta) {
  return new Promise(function(resolve, reject) {
    u2f.startAuthentication(appid, [meta])
    .then(function(authRequest) {
      resolve(authRequest);
    }, function(err) {
      reject(err);
    });
  });
}

function finishU2fAuthentication(u2f, authRequest, data, meta) {
  return new Promise(function(resolve, reject) {
    u2f.finishAuthentication(authRequest, data, [meta])
    .then(function(authenticationStatus) {
      resolve(authenticationStatus);
    }, function(err) {
      reject(err);
    })
  });
}

function sign_request(req, res) {
  var logger = req.app.get('logger');
  var user_data_storage = req.app.get('user data store');

  retrieveU2fMeta(req, user_data_storage)
  .then(function(doc) {
    if(!doc) {
      replyWithMissingRegistration(res);
      return;
    }

    var u2f = req.app.get('u2f');
    var meta = doc.meta;
    var appid = extractAppId(req);
    logger.info('U2F sign_request: Start authentication');
    return startU2fAuthentication(u2f, appid, meta);
  })
  .then(function(authRequest) {
    logger.info('U2F sign_request: Store authentication request and reply');
    req.session.auth_session.sign_request = authRequest;
    res.status(200);
    res.json(authRequest);
  })
  .catch(function(err) {
    logger.info('U2F sign_request: %s', err);
    replyWithUnauthorized(res);
  });
}


function sign(req, res) {
  if(!objectPath.has(req, 'session.auth_session.sign_request')) {
    replyWithUnauthorized(res); 
    return; 
  }

  var logger = req.app.get('logger');
  var user_data_storage = req.app.get('user data store');

  retrieveU2fMeta(req, user_data_storage)
  .then(function(doc) {
    var appid = extractAppId(req);
    var u2f = req.app.get('u2f');
    var authRequest = req.session.auth_session.sign_request;
    var meta = doc.meta;
    logger.info('U2F sign: Finish authentication');
    return finishU2fAuthentication(u2f, authRequest, req.body, meta);
  })
  .then(function(authenticationStatus) {
    logger.info('U2F sign: Authentication successful');
    req.session.auth_session.second_factor = true;
    res.status(204);
    res.send();
  })
  .catch(function(err) {
    logger.error('U2F sign: %s', err);
    res.status(401);
    res.send();
  });
}

