
var sinon = require('sinon');
var Promise = require('bluebird');
var assert = require('assert');
var first_factor = require('../../../src/lib/routes/first_factor');

describe('test the first factor validation route', function() {
  it('should return status code 204 when LDAP binding succeeds', function() {
    return test_first_factor_promised({ error: undefined, data: undefined }, 204);
  });

  it('should return status code 401 when LDAP binding fails', function() {
    return test_first_factor_promised({ error: 'ldap failed', data: undefined }, 401);
  });
});


function test_first_factor_promised(bind_params, statusCode) {
  return new Promise(function(resolve, reject) {
    test_first_factor(bind_params, statusCode, resolve, reject);
  });
}

function test_first_factor(bind_params, statusCode, resolve, reject) {
  var send = sinon.spy(function(data) {
    resolve();
  });
  var status = sinon.spy(function(code) {
    assert.equal(code, statusCode);
  });

  var bind_mock = sinon.stub().yields(bind_params.error, bind_params.data);
  var ldap_interface_mock = {
    bind: bind_mock
  }
  var config = {
    ldap_users_dn: 'dc=example,dc=com'
  }

  var app_get = sinon.stub();
  app_get.withArgs('ldap client').returns(ldap_interface_mock);
  app_get.withArgs('config').returns(ldap_interface_mock);
  var req = {
    app: {
      get: app_get
    },
    body: {
      username: 'username',
      password: 'password'
    }
  }
  var res = {
    send: send,
    status: status
  }

  first_factor(req, res);
}
