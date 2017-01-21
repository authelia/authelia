
var first_factor = require('./routes/first_factor');
var second_factor = require('./routes/second_factor');
var verify = require('./routes/verify');

module.exports = {
  login: serveLogin,
  logout: serveLogout,
  verify: verify,
  first_factor: first_factor,
  second_factor: second_factor
}

function serveLogin(req, res) {
  req.session.auth_session = {};
  req.session.auth_session.first_factor = false;
  req.session.auth_session.second_factor = false;

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

