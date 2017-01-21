
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

function run(config, ldap_client, u2f, fn) {
  var view_directory = path.resolve(__dirname, '../views');
  var public_html_directory = path.resolve(__dirname, '../public_html');

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
  app.set('config', config);
  
  app.get  ('/login',   routes.login);
  app.get  ('/logout',  routes.logout);

  app.get  ('/_verify',   routes.verify);

  app.post ('/_auth/1stfactor',        routes.first_factor);
  app.post ('/_auth/2ndfactor/totp',   routes.second_factor.totp);

  app.get  ('/_auth/2ndfactor/u2f/register_request',   routes.second_factor.u2f.register_request);
  app.post ('/_auth/2ndfactor/u2f/register',           routes.second_factor.u2f.register);
  app.get  ('/_auth/2ndfactor/u2f/sign_request',       routes.second_factor.u2f.sign_request);
  app.post ('/_auth/2ndfactor/u2f/sign',               routes.second_factor.u2f.sign);
  
  return app.listen(config.port, function(err) {
    console.log('Listening on %d...', config.port);
    if(fn) fn();
  });
}
