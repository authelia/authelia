#! /usr/bin/env node

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

var server = require('./lib/server');

var ldapjs = require('ldapjs');
var u2f = require('authdog');
var nodemailer = require('nodemailer');
var nedb = require('nedb');
var YAML = require('yamljs');
var session = require('express-session');
var winston = require('winston');
var speakeasy = require('speakeasy');

var config_path = process.argv[2];
if(!config_path) {
  console.log('No config file has been provided.');
  console.log('Usage: authelia <config>');
  process.exit(0);
}

console.log('Parse configuration file: %s', config_path);

var yaml_config = YAML.load(config_path);

var deps = {};
deps.u2f = u2f;
deps.nedb = nedb;
deps.nodemailer = nodemailer;
deps.ldapjs = ldapjs;
deps.session = session;
deps.winston = winston;
deps.speakeasy = speakeasy;

server.run(yaml_config, deps);
