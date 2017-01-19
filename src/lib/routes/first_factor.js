
module.exports = first_factor;

var ldap = require('../ldap');

function first_factor(req, res) {
  var username = req.body.username;
  var password = req.body.password;
  console.log('Start authentication of user %s', username);

  if(!username || !password) {
    replies.authentication_failed(res);
    return;
  }

  var ldap_client = req.app.get('ldap client');
  var config = req.app.get('config');

  ldap.validate(ldap_client, username, password, config.ldap_users_dn)
  .then(function() {
    res.status(204);
    res.send();
    console.log('LDAP binding successful');
  })
  .error(function(err) {
    res.status(401);
    res.send();
    console.log('LDAP binding failed:', err);
  });
}
