
module.exports = first_factor;

var ldap = require('../ldap');
var objectPath = require('object-path');

function replyWithUnauthorized(res) {
    res.status(401);
    res.send();
}

function first_factor(req, res) {
  if(!objectPath.has(req, 'session.auth_session.second_factor')) {
    replyWithUnauthorized(res);
  }
  
  var username = req.body.username;
  var password = req.body.password;
  console.log('Start authentication of user %s', username);

  if(!username || !password) {
    replyWithUnauthorized(res);
    return;
  }

  var ldap_client = req.app.get('ldap client');
  var config = req.app.get('config');

  ldap.validate(ldap_client, username, password, config.ldap_users_dn)
  .then(function() {
    req.session.auth_session.userid = username;
    req.session.auth_session.first_factor = true;
    res.status(204);
    res.send();
    console.log('LDAP binding successful');
  })
  .catch(function(err) {
    replyWithUnauthorized(res);
    console.log('LDAP binding failed:', err);
  });
}
