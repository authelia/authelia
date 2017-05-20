
var sinon = require('sinon');
var Promise = require('bluebird');
var assert = require('assert');
var winston = require('winston');
var first_factor = require('../../../src/lib/routes/first_factor');
var exceptions = require('../../../src/lib/exceptions');
var Ldap = require('../../../src/lib/ldap');

describe('test the first factor validation route', function() {
  var req, res;
  var ldap_interface_mock;
  var emails;
  var search_res_ok;
  var regulator;
  var access_controller;
  var config;

  beforeEach(function() {
    ldap_interface_mock = sinon.createStubInstance(Ldap);
    config = {
      ldap: {
        base_dn: 'ou=users,dc=example,dc=com',
        user_name_attribute: 'uid'
      }
    }

    emails = [ 'test_ok@example.com' ];
    groups = [ 'group1', 'group2' ];
 
    regulator = {};
    regulator.mark = sinon.stub();
    regulator.regulate = sinon.stub();

    regulator.mark.returns(Promise.resolve());
    regulator.regulate.returns(Promise.resolve());

    access_controller = {
      isDomainAllowedForUser: sinon.stub().returns(true)
    };

    var app_get = sinon.stub();
    app_get.withArgs('ldap').returns(ldap_interface_mock);
    app_get.withArgs('config').returns(config);
    app_get.withArgs('logger').returns(winston);
    app_get.withArgs('authentication regulator').returns(regulator);
    app_get.withArgs('access controller').returns(access_controller);

    req = {
      app: {
        get: app_get
      },
      body: {
        username: 'username',
        password: 'password'
      },
      session: {
        auth_session: {
          first_factor: false,
          second_factor: false
        }
      }
    }
    res = {
      send: sinon.spy(),
      status: sinon.spy()
    }
  });
  
  it('should return status code 204 when LDAP binding succeeds', function() {
    return new Promise(function(resolve, reject) {
      res.send = sinon.spy(function(data) {
        assert.equal('username', req.session.auth_session.userid);
        assert.equal(204, res.status.getCall(0).args[0]);
        resolve();
      });
      ldap_interface_mock.bind.withArgs('username').returns(Promise.resolve());
      ldap_interface_mock.get_emails.returns(Promise.resolve(emails));
      first_factor(req, res);
    });
  });

  it('should retrieve email from LDAP', function(done) {
    res.send = sinon.spy(function(data) { done(); });
    ldap_interface_mock.bind.returns(Promise.resolve());
    ldap_interface_mock.get_emails = sinon.stub().withArgs('usernam').returns(Promise.resolve([{mail: ['test@example.com'] }]));
    first_factor(req, res);
  });

  it('should set email as session variables', function() {
    return new Promise(function(resolve, reject) {
      res.send = sinon.spy(function(data) {
        assert.equal('test_ok@example.com', req.session.auth_session.email);
        resolve();
      });
      var emails = [ 'test_ok@example.com' ];
      ldap_interface_mock.bind.returns(Promise.resolve());
      ldap_interface_mock.get_emails.returns(Promise.resolve(emails));
      first_factor(req, res);
    });
  });

  it('should return status code 401 when LDAP binding throws', function(done) {
    res.send = sinon.spy(function(data) {
      assert.equal(401, res.status.getCall(0).args[0]);
      assert.equal(regulator.mark.getCall(0).args[0], 'username');
      done();
    });
    ldap_interface_mock.bind.throws(new exceptions.LdapBindError('Bad credentials'));
    first_factor(req, res);
  });

  it('should return status code 500 when LDAP search throws', function(done) {
    res.send = sinon.spy(function(data) {
      assert.equal(500, res.status.getCall(0).args[0]);
      done();
    });
    ldap_interface_mock.bind.returns(Promise.resolve());
    ldap_interface_mock.get_emails.throws(new exceptions.LdapSearchError('err'));
    first_factor(req, res);
  });

  it('should return status code 403 when regulator rejects authentication', function(done) {
    var err = new exceptions.AuthenticationRegulationError();
    regulator.regulate.returns(Promise.reject(err));
    res.send = sinon.spy(function(data) {
      assert.equal(403, res.status.getCall(0).args[0]);
      done();
    });
    ldap_interface_mock.bind.returns(Promise.resolve());
    ldap_interface_mock.get_emails.returns(Promise.resolve());
    first_factor(req, res);
  });
});


