
module.exports = {
  'validate': validateCredentials
}

var util = require('util');
var Promise = require('bluebird');

function validateCredentials(ldap_client, username, password, users_dn) {
  var userDN = util.format("binding entry cn=%s,%s", username, users_dn);
  var bind_promised = Promise.promisify(ldap_client.bind, ldap_client);
  return bind_promised(userDN, password);
}
