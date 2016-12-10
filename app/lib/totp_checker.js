
module.exports = {
  'validate': validate 
}

var Q = require('q');

function validate(speakeasy, token, totp_secret) {
  var defer = Q.defer(); 
  var real_token = speakeasy.totp({
    secret: totp_secret,
    encoding: 'base32'
  });

  if(token == real_token) {
    defer.resolve();
  }
  else {
    defer.reject('Wrong challenge');
  }
  return defer.promise;
}
