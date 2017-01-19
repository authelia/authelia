
module.exports = {
  'verify': verify_authentication
}

var objectPath = require('object-path');
var totp_checker = require('./totp_checker');
var replies = require('./replies');
var utils = require('./utils');

function verify_authentication(req, res) {
  console.log('Verify authentication');

  if(!objectPath.has(req, 'cookies.access_token')) {
    return utils.reject('No access token provided');
  }

  var jsonWebToken = req.cookies['access_token'];
  return req.app.get('jwt engine').verify(jsonWebToken);
}

