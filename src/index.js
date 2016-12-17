
var server = require('./lib/server');

var ldap = require('ldapjs');

var config = {
  port: process.env.PORT || 8080
  totp_secret: process.env.TOTP_SECRET,
  ldap_url: process.env.LDAP_URL || 'ldap://127.0.0.1:389',
  ldap_users_dn: process.env.LDAP_USERS_DN,
  jwt_secret: process.env.JWT_SECRET,
  jwt_expiration_time: process.env.JWT_EXPIRATION_TIME || '1h'
}

var ldap_client = ldap.createClient({
  url: config.ldap_url
});

server.run(config, ldap_client);

