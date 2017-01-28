
var sinon = require('sinon');
var Promise = require('bluebird');
var assert = require('assert');
var winston = require('winston');
var first_factor = require('../../../src/lib/routes/first_factor');
var exceptions = require('../../../src/lib/exceptions');

describe('test the first factor validation route', function() {
  var req, res;
  var ldap_interface_mock;
  var search_res_ok;
  var regulator;

  beforeEach(function() {
    ldap_interface_mock = {
      bind: sinon.stub(),
      search: sinon.stub()
    }
    var config = {
      ldap_users_dn: 'dc=example,dc=com'
    }

    var search_doc = {
      object: {
        mail: 'test_ok@example.com'
      }
    };
 
    var search_res_ok = {};
    search_res_ok.on = sinon.spy(function(event, fn) {
      if(event != 'error') fn(search_doc);
    });
    ldap_interface_mock.search.yields(undefined, search_res_ok);

    regulator = {};
    regulator.mark = sinon.stub();
    regulator.regulate = sinon.stub();

    regulator.mark.returns(Promise.resolve());
    regulator.regulate.returns(Promise.resolve());

    var app_get = sinon.stub();
    app_get.withArgs('ldap client').returns(ldap_interface_mock);
    app_get.withArgs('config').returns(ldap_interface_mock);
    app_get.withArgs('logger').returns(winston);
    app_get.withArgs('authentication regulator').returns(regulator);

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

  it('should return status code 401 when LDAP binding fails', function(done) {
    res.send = sinon.spy(function(data) {
      assert.equal(401, res.status.getCall(0).args[0]);
      assert.equal(regulator.mark.getCall(0).args[0], 'username');
      done();
    });
    ldap_interface_mock.bind.yields('Bad credentials');
    first_factor(req, res);
  });

  it('should return status code 500 when LDAP binding throws', function(done) {
    res.send = sinon.spy(function(data) {
      assert.equal(500, res.status.getCall(0).args[0]);
      done();
    });
    ldap_interface_mock.bind.yields(undefined);
    ldap_interface_mock.search.yields('error');
    first_factor(req, res);
  });

  it('should return status code 403 when regulator rejects authentication', function(done) {
    var err = new exceptions.AuthenticationRegulationError();
    regulator.regulate.returns(Promise.reject(err));
    res.send = sinon.spy(function(data) {
      assert.equal(403, res.status.getCall(0).args[0]);
      done();
    });
    ldap_interface_mock.bind.yields(undefined);
    ldap_interface_mock.search.yields(undefined);
    first_factor(req, res);
  });
});


