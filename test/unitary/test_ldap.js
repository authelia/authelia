
var ldap = require('../../src/lib/ldap');
var sinon = require('sinon');
var Promise = require('bluebird');
var assert = require('assert');


describe('test ldap validation', function() {
  var ldap_client;
  
  beforeEach(function() {
    ldap_client = {
      bind: sinon.stub(),
      search: sinon.stub(),
      modify: sinon.stub(),
      Change: sinon.spy()
    }
  });

  describe('test binding', test_binding);
  describe('test get email', test_get_email);
  describe('test update password', test_update_password);

  function test_binding() {
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
  }

  function test_get_email() {
    it('should retrieve the email of an existing user', function() {
      var expected_doc = {};
      expected_doc.object = {};
      expected_doc.object.mail = 'user@example.com';
      var res_emitter = {};
      res_emitter.on = sinon.spy(function(event, fn) {
        if(event != 'error') fn(expected_doc)
      });

      ldap_client.search.yields(undefined, res_emitter);

      return ldap.get_email(ldap_client, 'user', 'dc=example,dc=com')
      .then(function(doc) {
        assert.deepEqual(doc, expected_doc.object);
        return Promise.resolve(); 
      })
    });

    it('should fail on error with search method', function(done) {
      var expected_doc = {};
      expected_doc.mail = [];
      expected_doc.mail.push('user@example.com');
      ldap_client.search.yields('error');

      ldap.get_email(ldap_client, 'user', 'dc=example,dc=com')
      .catch(function() {
        done();
      })
    });
  }

  function test_update_password() {
    it('should update the password successfully', function(done) {
      var change = {};
      change.operation = 'replace';
      change.modification = {};
      change.modification.userPassword = 'new-password';

      var config = {};
      config.ldap_users_dn = 'dc=example,dc=com';
      config.ldap_user = 'admin';

      var userdn = 'cn=user,dc=example,dc=com';

      var ldapjs = {};
      ldapjs.Change = sinon.spy();

      ldap_client.bind.yields(undefined);
      ldap_client.modify.yields(undefined);

      ldap.update_password(ldap_client, ldapjs, 'user', 'new-password', config)
      .then(function() {
        assert.deepEqual(ldap_client.modify.getCall(0).args[0], userdn);
        assert.deepEqual(ldapjs.Change.getCall(0).args[0].operation, change.operation);
      
        var userPassword = ldapjs.Change.getCall(0).args[0].modification.userPassword;
        assert(/{SSHA}/.test(userPassword));
        done();
      })
    });

    it('should fail when ldap throws an error', function(done) {
      ldap_client.bind.yields(undefined);
      ldap_client.modify.yields('Error');

      var config = {};
      config.ldap_users_dn = 'dc=example,dc=com';
      config.ldap_user = 'admin';

      var ldapjs = {};
      ldapjs.Change = sinon.spy();

      ldap.update_password(ldap_client, ldapjs, 'user', 'new-password', config)
      .catch(function() {
        done();
      })
    });
  }
});

