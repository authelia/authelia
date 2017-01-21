
var denyNotLogged = require('./deny_not_logged');
var u2f = require('./u2f'); 

module.exports = {
  totp: denyNotLogged(require('./totp')),
  u2f: {
    register_request: denyNotLogged(u2f.register_request),
    register: denyNotLogged(u2f.register),
    sign_request: denyNotLogged(u2f.sign_request),
    sign: denyNotLogged(u2f.sign),
  }
}

