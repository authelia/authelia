var objectPath = require('object-path');
var Promise = require('bluebird');

var CHALLENGE = 'totp-register';

var icheck_interface = {
  challenge: CHALLENGE,
  render_template: 'totp-register',
  pre_check_callback: pre_check,
  email_subject: 'Register your TOTP secret key',
}

module.exports = {
  icheck_interface: icheck_interface,
  post: post,
}

function pre_check(req) {
  var first_factor_passed = objectPath.get(req, 'session.auth_session.first_factor');
  if(!first_factor_passed) {
    return Promise.reject('Authentication required before registering TOTP secret key');
  }

  var userid = objectPath.get(req, 'session.auth_session.userid');
  var email = objectPath.get(req, 'session.auth_session.email');

  if(!(userid && email)) {
    return Promise.reject('User ID or email is missing');
  }

  var identity = {};
  identity.email = email;
  identity.userid = userid;
  return Promise.resolve(identity);
}

// Generate a secret and send it to the user
function post(req, res) {
  var logger = req.app.get('logger');
  var userid = objectPath.get(req, 'session.auth_session.identity_check.userid');
  var challenge = objectPath.get(req, 'session.auth_session.identity_check.challenge');

  if(challenge != CHALLENGE || !userid) {
    res.status(403);
    res.send();
    return;
  }

  var user_data_store = req.app.get('user data store');
  var totp = req.app.get('totp engine');
  var secret = totp.generateSecret();

  logger.debug('POST new-totp-secret: save the TOTP secret in DB');
  user_data_store.set_totp_secret(userid, secret)
  .then(function() {
    var doc = {};
    doc.otpauth_url = secret.otpauth_url;
    doc.base32 = secret.base32;
    doc.ascii = secret.ascii;

    objectPath.set(req, 'session', undefined);

    res.status(200);
    res.json(doc);
  })
  .catch(function(err) {
    logger.error('POST new-totp-secret: Internal error %s', err);
    res.status(500);
    res.send();
  });
}

