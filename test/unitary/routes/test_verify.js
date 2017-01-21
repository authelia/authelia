
var assert = require('assert');
var verify = require('../../../src/lib/routes/verify');
var sinon = require('sinon');

describe('test authentication token verification', function() {
  var req, res;

  beforeEach(function() {
    req = {};
    res = {};
    res.status = sinon.spy();
  });

  it('should be already authenticated', function(done) {
    req.session = {};
    req.session.auth_session = {first_factor: true, second_factor: true};
 
    res.send = sinon.spy(function() {
      assert.equal(204, res.status.getCall(0).args[0]);
      done();
    });

    verify(req, res);
  });

  describe('given different cases of session', function() {
    function test_unauthorized(auth_session) {
      return new Promise(function(resolve, reject) {
        req.session = {};
        req.session.auth_session = auth_session;
 
        res.send = sinon.spy(function() {
          assert.equal(401, res.status.getCall(0).args[0]);
          resolve();
        });

        verify(req, res);
      });
    }

    it('should not be authenticated when second factor is missing', function() {
      return test_unauthorized({ first_factor: true, second_factor: false });
    });

    it('should not be authenticated when first factor is missing', function() {
      return test_unauthorized({ first_factor: false, second_factor: true });
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

