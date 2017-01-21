
var user_key_container = {};
var denyNotLogged = require('./deny_not_logged');
var u2f = require('./u2f')(user_key_container); // create a u2f handler bound to
// user key container

module.exports = {
  totp: denyNotLogged(require('./totp')),
  u2f: {
    register_request: denyNotLogged(u2f.register_request),
    register: denyNotLogged(u2f.register),
    sign_request: denyNotLogged(u2f.sign_request),
    sign: denyNotLogged(u2f.sign),
  }
}

