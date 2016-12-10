
module.exports = Jwt;

var jwt = require('jsonwebtoken');
var utils = require('./utils');
var Q = require('q');

function Jwt(secret) {
  var _secret;

  this._secret = secret;
}

Jwt.prototype.sign = function(data, expiration_time) {
  return jwt.sign(data, this._secret, { expiresIn: expiration_time });
}

Jwt.prototype.verify = function(token) {
  var defer = Q.defer();
  try {
    var decoded = jwt.verify(token, this._secret);
    defer.resolve(decoded);
  }
  catch(err) {
    defer.reject(err);
  }
  return defer.promise;
}

