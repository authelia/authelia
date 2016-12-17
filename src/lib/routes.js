
module.exports = {
  'auth': serveAuth,
  'login': serveLogin,
  'logout': serveLogout
}

var authentication = require('./authentication');
var replies = require('./replies');

function serveAuth(req, res) {
  if(req.method == 'POST') {
    serveAuthPost(req, res);
  }
  else {
    serveAuthGet(req, res);
  }
}

function serveAuthGet(req, res) {
  authentication.verify(req, res)
  .then(function(user) {
    replies.already_authenticated(res, user);
  })
  .fail(function(err) {
    replies.authentication_failed(res);
    console.error(err);
  });
}

function serveAuthPost(req, res) {
  authentication.authenticate(req, res);
}

function serveLogin(req, res) {
  console.log(req.headers);
  res.render('login');
}

function serveLogout(req, res) {
  var redirect_param = req.query.redirect;
  var redirect_url = redirect_param || '/';
  res.clearCookie('access_token');
  res.redirect(redirect_url);
}

