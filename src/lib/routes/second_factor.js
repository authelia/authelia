
var denyNotLogged = require('./deny_not_logged');
var u2f = require('./u2f'); 

module.exports = {
  totp: denyNotLogged(require('./totp')),
  u2f: {
    register_request: u2f.register_request,
    register: u2f.register,
    register_handler_get: u2f.register_handler_get,
    register_handler_post: u2f.register_handler_post,

    sign_request: denyNotLogged(u2f.sign_request),
    sign: denyNotLogged(u2f.sign),
  }
}

