
module.exports = totp;

var totp = require('../totp');
var objectPath = require('object-path');

var UNAUTHORIZED_MESSAGE = 'Unauthorized access';

function replyWithUnauthorized(res) {
    res.status(401);
    res.send();
}

function totp(req, res) {
  if(!objectPath.has(req, 'session.auth_session.second_factor')) {
    replyWithUnauthorized(res);
  }
  var token = req.body.token;
  
  var totp_engine = req.app.get('totp engine');
  var config = req.app.get('config');

  totp.validate(totp_engine, token, config.totp_secret)
  .then(function() {
    req.session.auth_session.second_factor = true;
    res.status(204);
    res.send();
  })
  .catch(function(err) {
    console.error(err);
    replyWithUnauthorized(res);
  });
}
