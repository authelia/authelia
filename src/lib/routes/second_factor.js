
var DenyNotLogged = require('./DenyNotLogged');
var u2f = require('./u2f');
var TOTPAuthenticator = require("./TOTPAuthenticator");

module.exports = {
  totp: DenyNotLogged(TOTPAuthenticator),
  u2f: {
    register_request: u2f.register_request,
    register: u2f.register,
    register_handler_get: u2f.register_handler_get,
    register_handler_post: u2f.register_handler_post,

    sign_request: DenyNotLogged(u2f.sign_request),
    sign: DenyNotLogged(u2f.sign),
  }
}

