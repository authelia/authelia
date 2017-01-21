
var ldap = require('../../src/lib/ldap');
var sinon = require('sinon');
var Promise = require('bluebird');
var assert = require('assert');


describe('test ldap validation', function() {
  var ldap_client;
  
  beforeEach(function() {
    ldap_client = {
      bind: sinon.stub()
    }
  });

  function test_validate() {
      var username = 'user';
      var password = 'password';
      var ldap_url = 'http://ldap';
      var users_dn = 'dc=example,dc=com';
      return ldap.validate(ldap_client, username, password, ldap_url, users_dn);
  }


  it('should bind the user if good credentials provided', function() {
    ldap_client.bind.yields();
    return test_validate();
  });

  // cover an issue with promisify context
  it('should promisify correctly', function() {
    function LdapClient() {
      this.test = 'abc';
    }
    LdapClient.prototype.bind = function(username, password, fn) {
      assert.equal('abc', this.test);
      fn();
    }
    ldap_client = new LdapClient();
    return test_validate();
  });

  it('should not bind the user if wrong credentials provided', function() {
    ldap_client.bind.yields('wrong credentials');
    var promise = test_validate();
    return promise.catch(function() {
      return Promise.resolve();
    });
  });
});

