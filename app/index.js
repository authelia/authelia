
var express = require('express');
var bodyParser = require('body-parser');
var cookieParser = require('cookie-parser');
var routes = require('./lib/routes');
var ldap = require('ldapjs');
var speakeasy = require('speakeasy');

var totpSecret = process.env.SECRET;
var LDAP_URL = process.env.LDAP_URL || 'ldap://127.0.0.1:389';
var USERS_DN = process.env.USERS_DN;
var PORT = process.env.PORT || 80
var JWT_SECRET = 'this is the secret';
var EXPIRATION_TIME = process.env.EXPIRATION_TIME || '1h';

var ldap_client = ldap.createClient({
  url: LDAP_URL
});


var app = express();
app.use(cookieParser());
app.use(express.static(__dirname + '/public_html'));
app.use(bodyParser.urlencoded({ extended: false }));

app.set('view engine', 'ejs');

app.get  ('/login',   routes.login);
app.post ('/login',   routes.login);

app.get  ('/logout',  routes.logout);
app.get  ('/_auth',   routes.auth);

app.listen(PORT, function(err) {
  console.log('Listening on %d...', PORT);
});

