
module.exports = {
  get: register_handler_get,
  post: register_handler_post
}

var objectPath = require('object-path');
var randomstring = require('randomstring');
var Promise = require('bluebird');
var util = require('util');

var u2f_common = require('./u2f_common');

function register_handler_get(req, res) {
  var logger = req.app.get('logger');
  logger.info('U2F register_handler: Continue registration process');

  var registration_token = objectPath.get(req, 'query.registration_token');
  logger.debug('U2F register_handler: registration_token=%s', registration_token);

  if(!registration_token) {
    res.status(403);
    res.send();
    return;
  }

  var user_data_store = req.app.get('user data store');

  logger.debug('U2F register_handler: verify token validity and consume it');
  user_data_store.consume_u2f_registration_token(registration_token)
  .then(function() {
    res.render('u2f_register');
  })
  .catch(function(err) {
    res.status(403);
    res.send();
  });
}

function send_u2f_registration_email(email_sender, original_url, email, token) {
  var url = util.format('%s?registration_token=%s', original_url, token); 
  var email_content = util.format('<a href="%s">Register</a>', url);
  return email_sender.send(email, 'U2F Registration', email_content);
}

function register_handler_post(req, res) {
  var logger = req.app.get('logger');
  logger.info('U2F register_handler: Starting registration process');
  logger.debug('U2F register_request: headers=%s', JSON.stringify(req.headers));

  var userid = objectPath.get(req, 'session.auth_session.userid');
  var email = objectPath.get(req, 'session.auth_session.email');
  var first_factor_passed = objectPath.get(req, 'session.auth_session.first_factor');
  
  // the user needs to have validated the first factor
  if(!(userid && first_factor_passed)) {
    var error = 'You need to be authenticated to register';
    logger.error('U2F register_handler: %s', error);
    res.status(403);
    res.send(error);
    return;
  }

  if(!email) {
    var error = util.format('No email has been found for user %s', userid);
    logger.error('U2F register_handler: %s', error);
    res.status(400);
    res.send(error);
    return;
  }

  var five_minutes = 4 * 60 * 1000;
  var user_data_store = req.app.get('user data store');
  var token = randomstring.generate({ length: 64 });

  logger.debug('U2F register_request: issue u2f registration token %s for 5 minutes', token);
  user_data_store.save_u2f_registration_token(userid, token, five_minutes)
  .then(function() {
    logger.debug('U2F register_request: Send u2f registration email to %s', email);
    var email_sender = req.app.get('email sender');
    var original_url = u2f_common.extract_original_url(req);
    return send_u2f_registration_email(email_sender, original_url, email, token);
  })
  .then(function() {
    res.status(204);
    res.send();
  })
  .catch(function(err) {
    logger.error('U2F register_handler: %s', err);
    res.status(500);
    res.send();
  });
}
