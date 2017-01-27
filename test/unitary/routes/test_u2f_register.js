var sinon = require('sinon');
var winston = require('winston');
var u2f_register = require('../../../src/lib/routes/u2f_register_handler');
var assert = require('assert');

describe('test register handler', function() {
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
    req.app.get.withArgs('user data store').returns(user_data_store);

    res = {};
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  });

  describe('test u2f registration check', test_registration_check);

  function test_registration_check() {
    it('should fail if first_factor has not been passed', function(done) {
      req.session.auth_session.first_factor = false;
      u2f_register.icheck_interface.pre_check_callback(req)
      .catch(function(err) {
        done();
      });
    });

    it('should fail if userid is missing', function(done) {
      req.session.auth_session.first_factor = false;
      req.session.auth_session.userid = undefined;

      u2f_register.icheck_interface.pre_check_callback(req)
      .catch(function(err) {
        done();
      });
    });

    it('should fail if email is missing', function(done) {
      req.session.auth_session.first_factor = false;
      req.session.auth_session.email = undefined;

      u2f_register.icheck_interface.pre_check_callback(req)
      .catch(function(err) {
        done();
      });
    });

    it('should succeed if first factor passed, userid and email are provided', function(done) {
      u2f_register.icheck_interface.pre_check_callback(req)
      .then(function(err) {
        done();
      });
    });
  }
});
