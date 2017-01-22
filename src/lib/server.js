
module.exports = {
  run: run
}

var routes = require('./routes');

var express = require('express');
var bodyParser = require('body-parser');
var speakeasy = require('speakeasy');
var path = require('path');
var session = require('express-session');
var winston = require('winston');
var DataStore = require('nedb');
var nodemailer = require('nodemailer');
var UserDataStore = require('./user_data_store');
var EmailSender = require('./email_sender');

function run(config, ldap_client, u2f, fn) {
  var view_directory = path.resolve(__dirname, '../views');
  var public_html_directory = path.resolve(__dirname, '../public_html');
  var datastore_options = {};
  datastore_options.directory = config.store_directory;

  var email_options = {};
  email_options.gmail = config.gmail;

  var app = express();
  app.use(express.static(public_html_directory));
  app.use(bodyParser.urlencoded({ extended: false }));
  app.use(bodyParser.json());
  app.set('trust proxy', 1); // trust first proxy

  app.use(session({
    secret: config.session_secret,
    resave: false,
    saveUninitialized: true,
    cookie: { 
      secure: false,
      maxAge: config.session_max_age
    },
  }));
  
  app.set('views', view_directory);
  app.set('view engine', 'ejs');

  winston.level = 'debug';

  app.set('logger', winston);
  app.set('ldap client', ldap_client);
  app.set('totp engine', speakeasy);
  app.set('u2f', u2f);
  app.set('user data store', new UserDataStore(DataStore, datastore_options));
  app.set('email sender', new EmailSender(nodemailer, email_options));
  app.set('config', config);
  
  // web pages
  app.get  ('/login',   routes.login);
  app.get  ('/logout',  routes.logout);

  app.get  ('/u2f-register', routes.second_factor.u2f.register_handler_get);
  app.post ('/u2f-register', routes.second_factor.u2f.register_handler_post);

  // verify authentication
  app.get  ('/verify',   routes.verify);

  // Authentication process
  app.post ('/1stfactor',        routes.first_factor);
  app.post ('/2ndfactor/totp',   routes.second_factor.totp);

  app.get  ('/2ndfactor/u2f/register_request',   routes.second_factor.u2f.register_request);
  app.post ('/2ndfactor/u2f/register',           routes.second_factor.u2f.register);

  app.get  ('/2ndfactor/u2f/sign_request',       routes.second_factor.u2f.sign_request);
  app.post ('/2ndfactor/u2f/sign',               routes.second_factor.u2f.sign);
  
  return app.listen(config.port, function(err) {
    console.log('Listening on %d...', config.port);
    if(fn) fn();
  });
}
