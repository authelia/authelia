
var Ldap = require('../../src/lib/ldap');
var sinon = require('sinon');
var Promise = require('bluebird');
var assert = require('assert');
var ldapjs = require('ldapjs');
var winston = require('winston');


describe('test ldap validation', function() {
  var ldap_client;
  var ldap, ldapjs;
  var ldap_config;
  
  beforeEach(function() {
    ldap_client = {
      bind: sinon.stub(),
      search: sinon.stub(),
      modify: sinon.stub(),
      on: sinon.stub()
    };

    ldapjs = {
      Change: sinon.spy(),
      createClient: sinon.spy(function() {
        return ldap_client;
      })
    }
    ldap_config = {
      url: 'http://localhost:324',
      user: 'admin',
      password: 'password',
      base_dn: 'dc=example,dc=com',
      additional_user_dn: 'ou=users'
    };

    var deps = {};
    deps.ldapjs = ldapjs;
    deps.winston = winston;

    ldap = new Ldap(deps, ldap_config);
    return ldap.connect();
  });

  describe('test binding', test_binding);
  describe('test get emails from username', test_get_emails);
  describe('test get groups from username', test_get_groups);
  describe('test update password', test_update_password);

  function test_binding() {
    function test_bind() {
      var username = "username";
      var password = "password";
      return ldap.bind(username, password);
    }

    it('should bind the user if good credentials provided', function() {
      ldap_client.bind.yields();
      return test_bind();
    });

    it('should bind the user with correct DN', function() {
      ldap_config.user_name_attribute = 'uid';
      var username = 'user';
      var password = 'password';
      ldap_client.bind.withArgs('uid=user,ou=users,dc=example,dc=com').yields(); 
      return ldap.bind(username, password);
    });

    it('should default to cn user search filter if no filter provided', function() {
      var username = 'user';
      var password = 'password';
      ldap_client.bind.withArgs('cn=user,ou=users,dc=example,dc=com').yields(); 
      return ldap.bind(username, password);
    });

    it('should not bind the user if wrong credentials provided', function() {
      ldap_client.bind.yields('wrong credentials');
      var promise = test_bind();
      return promise.catch(function() {
        return Promise.resolve();
      });
    });
  }

  function test_get_emails() {
    var res_emitter;
    var expected_doc;

    beforeEach(function() {
      expected_doc = {};
      expected_doc.object = {};
      expected_doc.object.mail = 'user@example.com';

      res_emitter = {};
      res_emitter.on = sinon.spy(function(event, fn) {
        if(event != 'error') fn(expected_doc)
      });
    });

    it('should retrieve the email of an existing user', function() {
      ldap_client.search.yields(undefined, res_emitter);

      return ldap.get_emails('user')
      .then(function(emails) {
        assert.deepEqual(emails, [expected_doc.object.mail]);
        return Promise.resolve(); 
      })
    });

    it('should retrieve email for user with uid name attribute', function() {
      ldap_config.user_name_attribute = 'uid';
      ldap_client.search.withArgs('uid=username,ou=users,dc=example,dc=com').yields(undefined, res_emitter);
      return ldap.get_emails('username')
      .then(function(emails) {
        assert.deepEqual(emails, ['user@example.com']);
        return Promise.resolve();
      });
    });

    it('should fail on error with search method', function() {
      var expected_doc = {};
      expected_doc.mail = [];
      expected_doc.mail.push('user@example.com');
      ldap_client.search.yields('error');

      return ldap.get_emails('user')
      .catch(function() {
        return Promise.resolve();
      })
    });
  }

  function test_get_groups() {
    var res_emitter;
    var expected_doc1, expected_doc2;

    beforeEach(function() {
      expected_doc1 = {};
      expected_doc1.object = {};
      expected_doc1.object.cn = 'group1';

      expected_doc2 = {};
      expected_doc2.object = {};
      expected_doc2.object.cn = 'group2';

      res_emitter = {};
      res_emitter.on = sinon.spy(function(event, fn) {
        if(event != 'error') fn(expected_doc1);
        if(event != 'error') fn(expected_doc2);
      });
    });

    it('should retrieve the groups of an existing user', function() {
      ldap_client.search.yields(undefined, res_emitter);
      return ldap.get_groups('user')
      .then(function(groups) {
        assert.deepEqual(groups, ['group1', 'group2']);
        return Promise.resolve(); 
      });
    });

    it('should reduce the scope to additional_group_dn', function(done) {
      ldap_config.additional_group_dn = 'ou=groups';
      ldap_client.search = sinon.spy(function(base_dn) {
        assert.equal(base_dn, 'ou=groups,dc=example,dc=com');
        done();
      });
      ldap.get_groups('user');
    });

    it('should use default group_name_attr if not provided', function(done) {
      ldap_client.search = sinon.spy(function(base_dn, query) {
        assert.equal(base_dn, 'dc=example,dc=com');
        assert.equal(query.filter, 'member=cn=user,ou=users,dc=example,dc=com');
        assert.deepEqual(query.attributes, ['cn']);
        done();
      });
      ldap.get_groups('user');
    });

    it('should fail on error with search method', function() {
      ldap_client.search.yields('error');
      return ldap.get_groups('user')
      .catch(function() {
        return Promise.resolve();
      })
    });
  }

  function test_update_password() {
    it('should update the password successfully', function() {
      var change = {};
      change.operation = 'replace';
      change.modification = {};
      change.modification.userPassword = 'new-password';

      var userdn = 'cn=user,ou=users,dc=example,dc=com';

      ldap_client.bind.yields(undefined);
      ldap_client.modify.yields(undefined);

      return ldap.update_password('user', 'new-password')
      .then(function() {
        assert.deepEqual(ldap_client.modify.getCall(0).args[0], userdn);
        assert.deepEqual(ldapjs.Change.getCall(0).args[0].operation, change.operation);
      
        var userPassword = ldapjs.Change.getCall(0).args[0].modification.userPassword;
        assert(/{SSHA}/.test(userPassword));
        return Promise.resolve();
      })
    });

    it('should fail when ldap throws an error', function() {
      ldap_client.bind.yields(undefined);
      ldap_client.modify.yields('Error');

      return ldap.update_password('user', 'new-password')
      .catch(function() {
        return Promise.resolve();
      })
    });

    it('should update password of user using particular user name attribute', function() {
      ldap_config.user_name_attribute = 'uid';

      ldap_client.bind.yields(undefined);
      ldap_client.modify.withArgs('uid=username,ou=users,dc=example,dc=com').yields();
      return ldap.update_password('username', 'newpass');
    });
  }
});

