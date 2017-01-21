
var server = require('../../src/lib/server');

var request = require('request');
var assert = require('assert');
var speakeasy = require('speakeasy');
var sinon = require('sinon');
var Promise = require('bluebird');
var tmp = require('tmp');

var request = Promise.promisifyAll(request);

var PORT = 8050;
var BASE_URL = 'http://localhost:' + PORT;

describe('test data persistence', function() {
  var u2f;
  var tmpDir;
  var ldap_client = {
    bind: sinon.stub()
  };
  var config;

  before(function() {
    u2f = {};
    u2f.startRegistration = sinon.stub();
    u2f.finishRegistration = sinon.stub();
    u2f.startAuthentication = sinon.stub();
    u2f.finishAuthentication = sinon.stub();

    ldap_client.bind.withArgs('cn=test_ok,ou=users,dc=example,dc=com', 
                              'password').yields(undefined);
    ldap_client.bind.withArgs('cn=test_nok,ou=users,dc=example,dc=com', 
                              'password').yields('error');
    tmpDir = tmp.dirSync({ unsafeCleanup: true });
    config = {
      port: PORT,
      totp_secret: 'totp_secret',
      ldap_url: 'ldap://127.0.0.1:389',
      ldap_users_dn: 'ou=users,dc=example,dc=com',
      session_secret: 'session_secret',
      session_max_age: 50000,
      store_directory: tmpDir.name
    };
  });

  after(function() {
    tmpDir.removeCallback();
  });

  it('should save a u2f meta and reload it after a restart of the server', function() {
    var server;
    var sign_request = {};
    var sign_status = {};
    var registration_request = {};
    var registration_status = {};
    u2f.startRegistration.returns(Promise.resolve(sign_request));
    u2f.finishRegistration.returns(Promise.resolve(sign_status));
    u2f.startAuthentication.returns(Promise.resolve(registration_request));
    u2f.finishAuthentication.returns(Promise.resolve(registration_status));
  
    var j1 = request.jar();
    var j2 = request.jar();
    return start_server(config, ldap_client, u2f)
    .then(function(s) {
      server = s;
      return execute_login(j1);
    })
    .then(function(res) {
      return execute_first_factor(j1);
    }) 
    .then(function() {
      return execute_u2f_registration(j1);
    })
    .then(function() {
      return execute_u2f_authentication(j1);
    })
    .then(function() {
      return stop_server(server);
    })
    .then(function() {
      return start_server(config, ldap_client, u2f)
    })
    .then(function(s) {
      server = s;
      return execute_login(j2);
    })
    .then(function() {
      return execute_first_factor(j2);
    }) 
    .then(function() {
      return execute_u2f_authentication(j2);
    })
    .then(function(res) {
      assert.equal(204, res.statusCode);
      server.close();
      return Promise.resolve();
    })
    .catch(function(err) {
      console.error(err);
      return Promise.reject(err);
    });
  });

  function start_server(config, ldap_client, u2f) {
    return new Promise(function(resolve, reject) {
      var s = server.run(config, ldap_client, u2f);
      resolve(s);
    });
  }

  function stop_server(s) {
    return new Promise(function(resolve, reject) {
      s.close();
      resolve();
    });
  }

  function execute_first_factor(jar) {
    return request.postAsync({ 
      url: BASE_URL + '/_auth/1stfactor',
      jar: jar,
      form: {
        username: 'test_ok',
        password: 'password'
      }
    });
  }

  function execute_u2f_registration(jar) {
    return request.getAsync({
      url: BASE_URL + '/_auth/2ndfactor/u2f/register_request',
      jar: jar
    })
    .then(function(res) {
      return request.postAsync({
        url: BASE_URL + '/_auth/2ndfactor/u2f/register',
        jar: jar,
        form: {
          s: 'test'
        }
      });
    });
  }

  function execute_u2f_authentication(jar) {
    return request.getAsync({
      url: BASE_URL + '/_auth/2ndfactor/u2f/sign_request',
      jar: jar
    })
    .then(function() {
      return request.postAsync({
        url: BASE_URL + '/_auth/2ndfactor/u2f/sign',
        jar: jar,
        form: {
          s: 'test'
        }
      });
    });
  }

  function execute_verification(jar) {
    return request.getAsync({ url: BASE_URL + '/_verify', jar: jar })
  }

  function execute_login(jar) {
    return request.getAsync({ url: BASE_URL + '/login', jar: jar })
  }
});
