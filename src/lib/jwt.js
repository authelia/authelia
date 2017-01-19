
module.exports = Jwt;

var jwt = require('jsonwebtoken');
var utils = require('./utils');
var Promise = require('bluebird');

function Jwt(secret) {
  this._secret = secret;
}

Jwt.prototype.sign = function(data, expiration_time) {
  var that = this;
  return new Promise(function(resolve, reject) {
    var token = jwt.sign(data, that._secret, { expiresIn: expiration_time })
    resolve(token);
  });
}

Jwt.prototype.verify = function(token) {
  var that = this;
  return new Promise(function(resolve, reject) {
    try {
      var decoded = jwt.verify(token, that._secret);
      resolve(decoded);
    }
    catch(err) {
      reject(err.message);
    }
  });
}

