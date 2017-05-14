var sinon = require('sinon');
var winston = require('winston');
var totp_register = require('../../../src/lib/routes/totp_register');
var assert = require('assert');
var Promise = require('bluebird');

describe('test totp register', function() {
  var req, res;
  var user_data_store;

  beforeEach(function() {
    req = {}
    req.app = {};
    req.app.get = sinon.stub();
    req.app.get.withArgs('logger').returns(winston);
    req.session = {};
    req.session.auth_session = {};
    req.session.auth_session.userid = 'user';
    req.session.auth_session.email = 'user@example.com';
    req.session.auth_session.first_factor = true;
    req.session.auth_session.second_factor = false;
    req.headers = {};
    req.headers.host = 'localhost';

    var options = {};
    options.inMemoryOnly = true;

    user_data_store = {};
    user_data_store.set_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    user_data_store.get_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    user_data_store.issue_identity_check_token = sinon.stub().returns(Promise.resolve({}));
    user_data_store.consume_identity_check_token = sinon.stub().returns(Promise.resolve({}));
    user_data_store.set_totp_secret = sinon.stub().returns(Promise.resolve({}));
    req.app.get.withArgs('user data store').returns(user_data_store);

    res = {};
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  });

  describe('test totp registration check', test_registration_check);
  describe('test totp post secret', test_post_secret);

  function test_registration_check() {
    it('should fail if first_factor has not been passed', function(done) {
      req.session.auth_session.first_factor = false;
      totp_register.icheck_interface.pre_check_callback(req)
      .catch(function(err) {
        done();
      });
    });

    it('should fail if userid is missing', function(done) {
      req.session.auth_session.first_factor = false;
      req.session.auth_session.userid = undefined;

      totp_register.icheck_interface.pre_check_callback(req)
      .catch(function(err) {
        done();
      });
    });

    it('should fail if email is missing', function(done) {
      req.session.auth_session.first_factor = false;
      req.session.auth_session.email = undefined;

      totp_register.icheck_interface.pre_check_callback(req)
      .catch(function(err) {
        done();
      });
    });

    it('should succeed if first factor passed, userid and email are provided', function(done) {
      totp_register.icheck_interface.pre_check_callback(req)
      .then(function(err) {
        done();
      });
    });
  }

  function test_post_secret() {
    it('should send the secret in json format', function(done) {
      req.app.get.withArgs('totp engine').returns(require('speakeasy'));
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.userid = 'user';
      req.session.auth_session.identity_check.challenge = 'totp-register';
      res.json = sinon.spy(function() {
        done();
      });
      totp_register.post(req, res);
    });

    it('should clear the session for reauthentication', function(done) {
      req.app.get.withArgs('totp engine').returns(require('speakeasy'));
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.userid = 'user';
      req.session.auth_session.identity_check.challenge = 'totp-register';
      res.json = sinon.spy(function() {
        assert.equal(req.session, undefined);
        done();
      });
      totp_register.post(req, res);
    });

    it('should return 403 if the identity check challenge is not set', function(done) {
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.challenge = undefined;
      res.send = sinon.spy(function() {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });
      totp_register.post(req, res);
    });

    it('should return 500 if db throws', function(done) {
      req.app.get.withArgs('totp engine').returns(require('speakeasy'));
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.userid = 'user';
      req.session.auth_session.identity_check.challenge = 'totp-register';
      user_data_store.set_totp_secret.returns(new Promise.reject('internal error'));

      res.send = sinon.spy(function() {
        assert.equal(res.status.getCall(0).args[0], 500);
        done();
      });
      totp_register.post(req, res);
    });
  }
});
