
module.exports = {
  'validate': validate 
}

var Promise = require('bluebird');

function validate(totp_engine, token, totp_secret) {
  return new Promise(function(resolve, reject) {
    var real_token = totp_engine.totp({
      secret: totp_secret,
      encoding: 'base32'
    });

    if(token == real_token) {
      resolve();
    }
    else {
      reject('Wrong challenge');
    }
  });
}
