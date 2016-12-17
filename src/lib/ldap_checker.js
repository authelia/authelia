
module.exports = {
  'validate': validateCredentials
}

var Q = require('q');
var util = require('util');
var utils = require('./utils');

function validateCredentials(ldap_client, username, password, users_dn) {
  var userDN = util.format("cn=%s,%s", username, users_dn);
  var bind_promised = utils.promisify(ldap_client.bind, ldap_client);
  return bind_promised(userDN, password);
}
