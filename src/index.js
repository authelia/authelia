
var server = require('./lib/server');

var ldap = require('ldapjs');
var u2f = require('authdog');

var config = {
  port: process.env.PORT || 8080,
  totp_secret: process.env.TOTP_SECRET,
  ldap_url: process.env.LDAP_URL || 'ldap://127.0.0.1:389',
  ldap_users_dn: process.env.LDAP_USERS_DN,
  session_secret: process.env.SESSION_SECRET,
  session_max_age: process.env.SESSION_MAX_AGE || 3600000 // in ms
}

var ldap_client = ldap.createClient({
  url: config.ldap_url,
  reconnect: true
});

server.run(config, ldap_client, u2f);
