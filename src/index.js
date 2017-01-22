
var server = require('./lib/server');

var ldap = require('ldapjs');
var u2f = require('authdog');
var YAML = require('yamljs');

var config_path = process.argv[2];
console.log('Parse configuration file: %s', config_path);

var yaml_config = YAML.load(config_path);

var config = {
  port: process.env.PORT || 8080,
  totp_secret: yaml_config.totp_secret,
  ldap_url: yaml_config.ldap.url || 'ldap://127.0.0.1:389',
  ldap_users_dn: yaml_config.ldap.base_dn,
  session_secret: yaml_config.session.secret,
  session_max_age: yaml_config.session.expiration || 3600000, // in ms
  store_directory: yaml_config.store_directory,
  debug_level: yaml_config.debug_level,
  gmail: {
    user: yaml_config.notifier.gmail.username,
    pass: yaml_config.notifier.gmail.password
  }
}

var ldap_client = ldap.createClient({
  url: config.ldap_url,
  reconnect: true
});

server.run(config, ldap_client, u2f);
