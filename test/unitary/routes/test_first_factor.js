
var sinon = require('sinon');
var Promise = require('bluebird');
var assert = require('assert');
var first_factor = require('../../../src/lib/routes/first_factor');

describe('test the first factor validation route', function() {
  var req, res;
  var ldap_interface_mock;

  beforeEach(function() {
    var bind_mock = sinon.stub();
    ldap_interface_mock = {
      bind: bind_mock
    }
    var config = {
      ldap_users_dn: 'dc=example,dc=com'
    }

    var app_get = sinon.stub();
    app_get.withArgs('ldap client').returns(ldap_interface_mock);
    app_get.withArgs('config').returns(ldap_interface_mock);
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
      ldap_interface_mock.bind.yields(undefined);
      first_factor(req, res);
    });
  });

  it('should return status code 401 when LDAP binding fails', function() {
    return new Promise(function(resolve, reject) {
      res.send = sinon.spy(function(data) {
        assert.equal(401, res.status.getCall(0).args[0]);
        resolve();
      });
      ldap_interface_mock.bind.yields('Bad credentials');
      first_factor(req, res);
    });
  });
});


