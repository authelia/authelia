
module.exports = denyNotLogged;

var objectPath = require('object-path');

function denyNotLogged(next) {
  return function(req, res) {
    var auth_session = req.session.auth_session;
    var first_factor = objectPath.has(req, 'session.auth_session.first_factor')
                         && req.session.auth_session.first_factor;
    if(!first_factor) {
      res.status(403);
      res.send();
      return;
    }

    next(req, res);
  }
}
