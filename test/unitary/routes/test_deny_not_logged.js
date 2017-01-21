
var sinon = require('sinon');
var Promise = require('bluebird');
var assert = require('assert');

var denyNotLogged = require('../../../src/lib/routes/deny_not_logged');

describe('test not logged', function() {
  it('should return status code 401 when auth_session has not been previously created', function() {
    return test_auth_session_not_created();
  });

  it('should return status code 401 when auth_session has failed first factor', function() {
    return test_auth_first_factor_not_validated();
  });

  it('should return status code 204 when auth_session has succeeded first factor stage', function() {
    return test_auth_with_first_factor_validated();
  });
});

function test_auth_session_not_created() {
  return new Promise(function(resolve, reject) {
    var send = sinon.spy(resolve);
    var status = sinon.spy(function(code) {
      assert.equal(401, code);
    });
    var req = {
      session: {}
    }

    var res = {
      send: send,
      status: status
    }

    denyNotLogged(reject)(req, res);
  });
}

function test_auth_first_factor_not_validated() {
  return new Promise(function(resolve, reject) {
    var send = sinon.spy(resolve);
    var status = sinon.spy(function(code) {
      assert.equal(401, code);
    });
    var req = {
      session: {
        auth_session: {
          first_factor: false,
          second_factor: false
        }
      }
    }

    var res = {
      send: send,
      status: status
    }

    denyNotLogged(reject)(req, res);
  });
}

function test_auth_with_first_factor_validated() {
  return new Promise(function(resolve, reject) {
    var req = {
      session: {
        auth_session: {
          first_factor: true,
          second_factor: false
        }
      }
    }

    var res = {
      send: sinon.spy(),
      status: sinon.spy()
    }

    denyNotLogged(resolve)(req, res);
  });
}
