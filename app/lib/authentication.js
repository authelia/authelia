
module.exports = {
  'authenticate': authenticate,
  'verify_authentication': verify_authentication
}

var objectPath = require('object-path');
var Jwt = require('./jwt');
var ldap_checker = require('./ldap_checker');
var totp_checker = require('./totp_checker');
var replies = require('./replies');
var Q = require('q');
var utils = require('./utils');


function authenticate(req, res, args) {
  var defer = Q.defer();
  var username = req.body.username;
  var password = req.body.password;
  var token = req.body.token;
  console.log('Start authentication');

  if(!username || !password || !token) {
    replies.authentication_failed(res);
    return;
  }

  var totp_promise = totp_checker.validate(args.totp_interface, token, args.totp_secret);
  var credentials_promise = ldap_checker.validate(args.ldap_interface, username, password, args.users_dn);

  Q.all([totp_promise, credentials_promise])
  .then(function() {
    var token = args.jwt.sign({ user: username }, args.jwt_expiration_time);
    res.cookie('access_token', token);
    res.redirect('/');
    console.log('Authentication succeeded');
    defer.resolve();
  })
  .fail(function(err1, err2) {
    res.render('login');
    console.log('Authentication failed', err1, err2);
    defer.reject();
  });
  return defer.promise;
}

function verify_authentication(req, res, args) {
  console.log('Verify authentication');

  if(!objectPath.has(req, 'cookies.access_token')) {
    return utils.reject('No access token provided');
  }

  var jsonWebToken = req.cookies['access_token'];
  return args.jwt.verify(jsonWebToken);
}

