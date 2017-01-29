var objectPath = require('object-path');
var Promise = require('bluebird');
var QRCode = require('qrcode');

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


function secretToDataURLAsync(secret) {
  return new Promise(function(resolve, reject) {
    QRCode.toDataURL(secret.otpauth_url, function(err, url_data) {
      if(err) {
        reject(err);
        return;
      }
      resolve(url_data);
    });
  });
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
  var qrcode_data;

  secretToDataURLAsync(secret)
  .then(function(data) {
    qrcode_data = data;
    logger.debug('POST new-totp-secret: save the TOTP secret in DB');
    return user_data_store.set_totp_secret(userid, secret);
  })
  .then(function() {
    var doc = {};
    doc.qrcode = qrcode_data;
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

