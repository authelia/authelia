
var first_factor = require('./routes/first_factor');

module.exports = {
  auth: serveAuth,
  login: serveLogin,
  logout: serveLogout,
  first_factor: first_factor
}

var authentication = require('./authentication');
var replies = require('./replies');

function serveAuth(req, res) {
  serveAuthGet(req, res);
}

function serveAuthGet(req, res) {
  authentication.verify(req, res)
  .then(function(user) {
    replies.already_authenticated(res, user);
  })
  .catch(function(err) {
    replies.authentication_failed(res);
    console.error(err);
  });
}

function serveLogin(req, res) {
  res.render('login');
}

function serveLogout(req, res) {
  var redirect_param = req.query.redirect;
  var redirect_url = redirect_param || '/';
  res.clearCookie('access_token');
  res.redirect(redirect_url);
}

