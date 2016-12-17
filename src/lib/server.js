
module.exports = {
  run: run
}

var routes = require('./routes');
var Jwt = require('./jwt');

var express = require('express');
var bodyParser = require('body-parser');
var cookieParser = require('cookie-parser');
var speakeasy = require('speakeasy');
var path = require('path');

function run(config, ldap_client) {
  var view_directory = path.resolve(__dirname, '../views');
  var public_html_directory = path.resolve(__dirname, '../public_html');

  var app = express();
  app.use(cookieParser());
  app.use(express.static(public_html_directory));
  app.use(bodyParser.urlencoded({ extended: false }));
  
  app.set('views', view_directory);
  app.set('view engine', 'ejs');

  app.set('jwt engine', new Jwt(config.jwt_secret));
  app.set('ldap client', ldap_client);
  app.set('totp engine', speakeasy);
  app.set('config', config);
  
  app.get  ('/login',   routes.login);
  app.get  ('/logout',  routes.logout);

  app.get  ('/_auth',   routes.auth);
  app.post ('/_auth',   routes.auth);
  
  app.listen(config.port, function(err) {
    console.log('Listening on %d...', config.port);
  });
}
