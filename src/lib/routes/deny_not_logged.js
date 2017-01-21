
module.exports = denyNotLogged;

var objectPath = require('object-path');

function replyWithUnauthorized(res) {
  res.status(401);
  res.send('Unauthorized access');
}

function denyNotLogged(next) {
  return function(req, res) {
    var auth_session = req.session.auth_session;
    var first_factor = objectPath.has(req, 'session.auth_session.first_factor')
                         && req.session.auth_session.first_factor;
    if(!first_factor) {
      replyWithUnauthorized(res);
      console.log('Access to this route is denied');
      return;
    }

    next(req, res);
  }
}
