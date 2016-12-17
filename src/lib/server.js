
module.exports = {
  run: run
}

var routes = require('./routes');
var Jwt = require('./jwt');

var express = require('express');
var bodyParser = require('body-parser');
var cookieParser = require('cookie-parser');
var speakeasy = require('speakeasy');

function run(config, ldap_client) {
  var app = express();
  app.set('views', './src/views');
  app.use(cookieParser());
  app.use(express.static(__dirname + '/public_html'));
  app.use(bodyParser.urlencoded({ extended: false }));
  
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
