
module.exports = {
  'authenticate': authenticate,
  'verify': verify_authentication
}

var objectPath = require('object-path');
var ldap_checker = require('./ldap_checker');
var totp_checker = require('./totp_checker');
var replies = require('./replies');
var Q = require('q');
var utils = require('./utils');


function authenticate(req, res) {
  var defer = Q.defer();
  var username = req.body.username;
  var password = req.body.password;
  var token = req.body.token;
  console.log('Start authentication of user %s', username);

  if(!username || !password || !token) {
    replies.authentication_failed(res);
    return;
  }

  var jwt_engine = req.app.get('jwt engine');
  var ldap_client = req.app.get('ldap client');
  var totp_engine = req.app.get('totp engine');
  var config = req.app.get('config');

  var totp_promise = totp_checker.validate(totp_engine, token, config.totp_secret);
  var credentials_promise = ldap_checker.validate(ldap_client, username, password, config.ldap_users_dn);

  Q.all([totp_promise, credentials_promise])
  .then(function() {
    var token = jwt_engine.sign({ user: username }, config.jwt_expiration_time);
    replies.authentication_succeeded(res, username, token);
    console.log('Authentication succeeded');
    defer.resolve();
  })
  .fail(function(err1, err2) {
    console.log('Authentication failed', err1, err2);
    replies.authentication_failed(res);
    defer.reject();
  });
  return defer.promise;
}

function verify_authentication(req, res) {
  console.log('Verify authentication');

  if(!objectPath.has(req, 'cookies.access_token')) {
    return utils.reject('No access token provided');
  }

  var jsonWebToken = req.cookies['access_token'];
  return req.app.get('jwt engine').verify(jsonWebToken);
}

