#! /usr/bin/env node

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

var server = require('./lib/server');

var ldap = require('ldapjs');
var u2f = require('authdog');
var nodemailer = require('nodemailer');
var nedb = require('nedb');
var YAML = require('yamljs');
var session = require('express-session');

var config_path = process.argv[2];
if(!config_path) {
  console.log('No config file has been provided.');
  console.log('Usage: authelia <config>');
  process.exit(0);
}

console.log('Parse configuration file: %s', config_path);

var yaml_config = YAML.load(config_path);

var config = {
  port: process.env.PORT || 8080,
  ldap_url: yaml_config.ldap.url || 'ldap://127.0.0.1:389',
  ldap_user_search_base: yaml_config.ldap.user_search_base,
  ldap_user_search_filter: yaml_config.ldap.user_search_filter,
  ldap_user: yaml_config.ldap.user,
  ldap_password: yaml_config.ldap.password,
  session_domain: yaml_config.session.domain,
  session_secret: yaml_config.session.secret,
  session_max_age: yaml_config.session.expiration || 3600000, // in ms
  store_directory: yaml_config.store_directory,
  logs_level: yaml_config.logs_level,
  notifier: yaml_config.notifier,
}

var ldap_client = ldap.createClient({
  url: config.ldap_url,
  reconnect: true
});

ldap_client.on('error', function(err) {
  console.error('LDAP Error:', err.message)
})

var deps = {};
deps.u2f = u2f;
deps.nedb = nedb;
deps.nodemailer = nodemailer;
deps.ldap = ldap;
deps.session = session;

server.run(config, ldap_client, deps);
