
module.exports = {
  'auth': serveAuth,
  'login': serveLogin,
  'logout': serveLogout
}

var authentication = require('./authentication');
var replies = require('./replies');

function serveAuth(req, res) {
  authentication.verify(req, res)
  .then(function(user) {
    replies.already_authenticated(res, user);
  })
  .fail(function(err) {
    replies.authentication_failed(res);
    console.error(err);
  });
}

function serveLogin(req, res) {
  console.log('METHOD=%s', req.method);
  if(req.method == 'POST') {
    authentication.authenticate(req, res);
  }
  else {
    res.render('login');
  }
}

function serveLogout(req, res) {
  res.clearCookie('access_token');
  res.redirect('/');
}
