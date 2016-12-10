
var ldap_checker = require('../lib/ldap_checker');
var sinon = require('sinon');
var sinonPromise = require('sinon-promise');

sinonPromise(sinon);

var autoResolving = sinon.promise().resolves();

function test_validate(bind_mock) {
    var username = 'user';
    var password = 'password';
    var ldap_url = 'http://ldap';
    var users_dn = 'dc=example,dc=com';
    
    var ldap_client_mock = {
      bind: bind_mock
    }

    return ldap_checker.validate(ldap_client_mock, username, password, ldap_url, users_dn);
}

describe('test ldap checker', function() {
  it('should bind the user if good credentials provided', function() {
    var bind_mock = sinon.mock().yields();
    return test_validate(bind_mock);
  });

  it('should not bind the user if wrong credentials provided', function() {
    var bind_mock = sinon.mock().yields('wrong credentials');
    var promise = test_validate(bind_mock);
    return promise.fail(autoResolving);
  });
});

