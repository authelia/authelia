
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
var UserDataStore = require('./user_data_store');
var Notifier = require('./notifier');
var AuthenticationRegulator = require('./authentication_regulator');
var identity_check = require('./identity_check');

function run(config, ldap_client, deps, fn) {
  var view_directory = path.resolve(__dirname, '../views');
  var public_html_directory = path.resolve(__dirname, '../public_html');
  var datastore_options = {};
  datastore_options.directory = config.store_directory;
  if(config.store_in_memory)
    datastore_options.inMemory = true;

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

  // by default the level of logs is info
  winston.level = config.logs_level || 'info';

  var five_minutes = 5 * 60;
  var data_store = new UserDataStore(deps.nedb, datastore_options);
  var regulator = new AuthenticationRegulator(data_store, five_minutes);
  var notifier = new Notifier(config.notifier, deps);

  app.set('logger', winston);
  app.set('ldap', deps.ldap);
  app.set('ldap client', ldap_client);
  app.set('totp engine', speakeasy);
  app.set('u2f', deps.u2f);
  app.set('user data store', data_store);
  app.set('notifier', notifier);
  app.set('authentication regulator', regulator);
  app.set('config', config);

  var base_endpoint = '/authentication';
  
  // web pages
  app.get  (base_endpoint + '/login',   routes.login);
  app.get  (base_endpoint + '/logout',  routes.logout);

  identity_check(app, base_endpoint + '/totp-register', routes.totp_register.icheck_interface);
  identity_check(app, base_endpoint + '/u2f-register', routes.u2f_register.icheck_interface);
  identity_check(app, base_endpoint + '/reset-password', routes.reset_password.icheck_interface);

  app.get  (base_endpoint + '/reset-password-form', function(req, res) { res.render('reset-password-form'); });

  // Reset the password
  app.post (base_endpoint + '/new-password', routes.reset_password.post);

  // Generate a new TOTP secret
  app.post (base_endpoint + '/new-totp-secret', routes.totp_register.post);

  // verify authentication
  app.get  (base_endpoint + '/verify',   routes.verify);

  // Authentication process
  app.post (base_endpoint + '/1stfactor',        routes.first_factor);
  app.post (base_endpoint + '/2ndfactor/totp',   routes.second_factor.totp);

  // U2F registration
  app.get  (base_endpoint + '/2ndfactor/u2f/register_request',   routes.second_factor.u2f.register_request);
  app.post (base_endpoint + '/2ndfactor/u2f/register',           routes.second_factor.u2f.register);

  // U2F authentication
  app.get  (base_endpoint + '/2ndfactor/u2f/sign_request',       routes.second_factor.u2f.sign_request);
  app.post (base_endpoint + '/2ndfactor/u2f/sign',               routes.second_factor.u2f.sign);
  
  return app.listen(config.port, function(err) {
    console.log('Listening on %d...', config.port);
    if(fn) fn();
  });
}
