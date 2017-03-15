
var u2f_register = require('./u2f_register');
var u2f_common = require('./u2f_common');
var objectPath = require('object-path');

module.exports = {
  register_request: u2f_register.register_request,
  register: u2f_register.register,
  register_handler_get: u2f_register.register_handler_get,
  register_handler_post: u2f_register.register_handler_post,

  sign_request: sign_request,
  sign: sign,
}


function retrieve_u2f_meta(req, user_data_storage) {
  var userid = req.session.auth_session.userid;
  var appid = u2f_common.extract_app_id(req);
  return user_data_storage.get_u2f_meta(userid, appid);
}


function sign_request(req, res) {
  var logger = req.app.get('logger');
  var user_data_storage = req.app.get('user data store');

  retrieve_u2f_meta(req, user_data_storage)
  .then(function(doc) {
    if(!doc) {
      u2f_common.reply_with_missing_registration(res);
      return;
    }

    var u2f = req.app.get('u2f');
    var meta = doc.meta;
    var appid = u2f_common.extract_app_id(req);
    logger.info('U2F sign_request: Start authentication to app %s', appid);
    return u2f.startAuthentication(appid, [meta])
  })
  .then(function(authRequest) {
    logger.info('U2F sign_request: Store authentication request and reply');
    req.session.auth_session.sign_request = authRequest;
    res.status(200);
    res.json(authRequest);
  })
  .catch(function(err) {
    logger.info('U2F sign_request: %s', err);
    res.status(500);
    res.send();
  });
}


function sign(req, res) {
  if(!objectPath.has(req, 'session.auth_session.sign_request')) {
    u2f_common.reply_with_unauthorized(res); 
    return; 
  }

  var logger = req.app.get('logger');
  var user_data_storage = req.app.get('user data store');

  retrieve_u2f_meta(req, user_data_storage)
  .then(function(doc) {
    var appid = u2f_common.extract_app_id(req);
    var u2f = req.app.get('u2f');
    var authRequest = req.session.auth_session.sign_request;
    var meta = doc.meta;
    logger.info('U2F sign: Finish authentication');
    return u2f.finishAuthentication(authRequest, req.body, [meta])
  })
  .then(function(authenticationStatus) {
    logger.info('U2F sign: Authentication successful');
    req.session.auth_session.second_factor = true;
    res.status(204);
    res.send();
  })
  .catch(function(err) {
    logger.error('U2F sign: %s', err);
    res.status(500);
    res.send();
  });
}

