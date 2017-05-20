
var assert = require('assert');
var verify = require('../../../src/lib/routes/verify');
var sinon = require('sinon');
var winston = require('winston');

describe('test authentication token verification', function() {
  var req, res;
  var config_mock;
  var acl_matcher;

  beforeEach(function() {
    acl_matcher = {
      is_domain_allowed: sinon.stub().returns(true)
    };
    var access_control = {
      matcher: acl_matcher
    }
    config_mock = {};
    req = {};
    res = {};
    req.headers = {};
    req.headers.host = 'secret.example.com';
    req.app = {};
    req.app.get = sinon.stub();
    req.app.get.withArgs('config').returns(config_mock);
    req.app.get.withArgs('logger').returns(winston);
    req.app.get.withArgs('access control').returns(access_control);
    res.status = sinon.spy();
  });

  it('should be already authenticated', function(done) {
    req.session = {};
    req.session.auth_session = {
      first_factor: true, 
      second_factor: true,
      userid: 'myuser',
      allowed_domains: ['*']
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

    it('should not be authenticated when session is partially initialized', function() {
      return test_unauthorized({ first_factor: true });
    });
  });
});

