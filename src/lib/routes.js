
var first_factor = require('./routes/FirstFactor');
var second_factor = require('./routes/second_factor');
var reset_password = require('./routes/reset_password');
var verify = require('./routes/verify');
var u2f_register_handler = require('./routes/u2f_register_handler');
var totp_register = require('./routes/totp_register');
var objectPath = require('object-path');

module.exports = {
  login: serveLogin,
  logout: serveLogout,
  verify: verify,
  first_factor: first_factor,
  second_factor: second_factor,
  reset_password: reset_password,
  u2f_register: u2f_register_handler,
  totp_register: totp_register,
}

function serveLogin(req, res) {
  if(!(objectPath.has(req, 'session.auth_session'))) {
    req.session.auth_session = {};
    req.session.auth_session.first_factor = false;
    req.session.auth_session.second_factor = false;
  }
  res.render('login');
}

function serveLogout(req, res) {
  var redirect_param = req.query.redirect;
  var redirect_url = redirect_param || '/';
  req.session.auth_session = {
    first_factor: false,
    second_factor: false
  }
  res.redirect(redirect_url);
}

