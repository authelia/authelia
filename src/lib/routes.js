
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
  res.render('login');
}

function serveLogout(req, res) {
  res.clearCookie('access_token');
  res.redirect('/');
}
