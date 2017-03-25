
var assert = require('assert');
var verify = require('../../../src/lib/routes/verify');
var sinon = require('sinon');
var winston = require('winston');

describe('test authentication token verification', function() {
  var req, res;
  var config_mock;

  beforeEach(function() {
    config_mock = {};
    req = {};
    res = {};
    req.headers = {};
    req.headers.host = 'secret.example.com';
    req.app = {};
    req.app.get = sinon.stub();
    req.app.get.withArgs('config').returns(config_mock);
    req.app.get.withArgs('logger').returns(winston);
    res.status = sinon.spy();
  });

  it('should be already authenticated', function(done) {
    req.session = {};
    req.session.auth_session = {
      first_factor: true, 
      second_factor: true,
      userid: 'myuser',
      group: 'mygroup'
    };
 
    res.send = sinon.spy(function() {
      assert.equal(204, res.status.getCall(0).args[0]);
      done();
    });

    verify(req, res);
  });

  describe('given different cases of session', function() {
    function test_session(auth_session, status_code) {
      return new Promise(function(resolve, reject) {
        req.session = {};
        req.session.auth_session = auth_session;
 
        res.send = sinon.spy(function() {
          assert.equal(status_code, res.status.getCall(0).args[0]);
          resolve();
        });

        verify(req, res);
      });
    }

    function test_unauthorized(auth_session) {
      return test_session(auth_session, 401);
    }

    function test_authorized(auth_session) {
      return test_session(auth_session, 204);
    }

    it('should not be authenticated when second factor is missing', function() {
      return test_unauthorized({ first_factor: true, second_factor: false });
    });

    it('should not be authenticated when first factor is missing', function() {
      return test_unauthorized({ first_factor: false, second_factor: true });
    });

    it('should not be authenticated when userid is missing', function() {
      return test_unauthorized({ 
        first_factor: true, 
        second_factor: true,
        group: 'mygroup',
      });
    });

    it('should not be authenticated when first and second factor are missing', function() {
      return test_unauthorized({ first_factor: false, second_factor: false });
    });

    it('should not be authenticated when session has not be initiated', function() {
      return test_unauthorized(undefined);
    });

    it('should reply unauthorized when the domain is restricted', function() {
      config_mock.access_control = [];
      config_mock.access_control.push({
        group: 'abc',
        allowed_domains: ['secret.example.com']
      });
      return test_unauthorized({
        first_factor: true,
        second_factor: true,
        userid: 'user',
        allowed_domains: ['restricted.example.com']
      });
    });

    it('should reply authorized when the domain is allowed', function() {
      config_mock.access_control = [];
      config_mock.access_control.push({
        group: 'abc',
        allowed_domains: ['secret.example.com']
      });
      return test_authorized({
        first_factor: true,
        second_factor: true,
        userid: 'user',
        allowed_domains: ['secret.example.com']
      });
    });

    it('should not be authenticated when session is partially initialized', function() {
      return test_unauthorized({ first_factor: true });
    });
  });
});

