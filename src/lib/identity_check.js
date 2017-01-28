
var objectPath = require('object-path');
var randomstring = require('randomstring');
var Promise = require('bluebird');
var util = require('util');
var exceptions = require('./exceptions');
var fs = require('fs');
var ejs = require('ejs');

module.exports = identity_check;

var filePath = __dirname + '/../resources/email-template.ejs';
var email_template = fs.readFileSync(filePath, 'utf8');

// IdentityCheck class

function IdentityCheck(user_data_store, email_sender, logger) {
  this._user_data_store = user_data_store;
  this._email_sender = email_sender;
  this._logger = logger;
}

IdentityCheck.prototype.issue_token = function(userid, email, content, logger) {
  var five_minutes = 4 * 60 * 1000;
  var token = randomstring.generate({ length: 64 });
  var that = this;

  this._logger.debug('identity_check: issue identity token %s for 5 minutes', token);
  return this._user_data_store.issue_identity_check_token(userid, token, content, five_minutes)
  .then(function() {
    return Promise.resolve(token);
  });
}

IdentityCheck.prototype.consume_token = function(token, logger) {
  this._logger.debug('identity_check: consume token %s', token);
  return this._user_data_store.consume_identity_check_token(token)
}


// The identity_check middleware that allows the user two perform a two step validation
// using the user email

function identity_check(app, endpoint, icheck_interface) {
  app.get(endpoint, identity_check_get(endpoint, icheck_interface)); 
  app.post(endpoint, identity_check_post(endpoint, icheck_interface)); 
}


function identity_check_get(endpoint, icheck_interface) {
  return function(req, res) {
    var logger = req.app.get('logger');
    var identity_token = objectPath.get(req, 'query.identity_token');
    logger.info('GET identity_check: identity token provided is %s', identity_token);

    if(!identity_token) {
      res.status(403);
      res.send();
      return;
    }
 
    var email_sender = req.app.get('email sender');
    var user_data_store = req.app.get('user data store');
    var identity_check = new IdentityCheck(user_data_store, email_sender, logger);

    identity_check.consume_token(identity_token, logger)
    .then(function(content) {
      objectPath.set(req, 'session.auth_session.identity_check', {});
      req.session.auth_session.identity_check.challenge = icheck_interface.challenge;
      req.session.auth_session.identity_check.userid = content.userid;
      res.render(icheck_interface.render_template);
    }, function(err) {
      logger.error('GET identity_check: Error while consuming token %s', err);
      throw new exceptions.AccessDeniedError('Access denied');
    })
    .catch(exceptions.AccessDeniedError, function(err) {
      logger.error('GET identity_check: Access Denied %s', err);
      res.status(403);
      res.send();
    })
    .catch(function(err) {
      logger.error('GET identity_check: Internal error %s', err);
      res.status(500);
      res.send();
    });
  }
}


function identity_check_post(endpoint, icheck_interface) {
  return function(req, res) {
    var logger = req.app.get('logger');
    var email_sender = req.app.get('email sender');
    var user_data_store = req.app.get('user data store');
    var identity_check = new IdentityCheck(user_data_store, email_sender, logger);
    var userid, email_address;

    icheck_interface.pre_check_callback(req)
    .then(function(identity) {
      email_address = objectPath.get(identity, 'email');
      userid = objectPath.get(identity, 'userid');

      if(!(email_address && userid)) {
        throw new exceptions.IdentityError('Missing user id or email address');
      }

      return identity_check.issue_token(userid, undefined, logger);
    }, function(err) {
      throw new exceptions.AccessDeniedError();
    })
    .then(function(token) {
      var original_url = util.format('https://%s%s', req.headers.host, req.headers['x-original-uri']);
      var link_url = util.format('%s?identity_token=%s', original_url, token); 
      var email = {};

      var d = {};
      d.url = link_url;
      d.button_title = 'Continue';
      d.title = icheck_interface.email_subject;

      email.to = email_address;
      email.subject = icheck_interface.email_subject;
      email.content = ejs.render(email_template, d);

      logger.info('POST identity_check: send email to %s', email.to);
      return email_sender.send(email.to, email.subject, email.content);
    })
    .then(function() {
      res.status(204);
      res.send();
    })
    .catch(exceptions.IdentityError, function(err) {
      logger.error('POST identity_check: IdentityError %s', err);
      res.status(400);
      res.send();
    })
    .catch(exceptions.AccessDeniedError, function(err) {
      logger.error('POST identity_check: AccessDeniedError %s', err);
      res.status(403);
      res.send();
    })
    .catch(function(err) {
      logger.error('POST identity_check: Error %s', err);
      res.status(500);
      res.send();
    });
  }
}


