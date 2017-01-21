
var server = require('../../src/lib/server');

var request = require('request');
var assert = require('assert');
var speakeasy = require('speakeasy');
var sinon = require('sinon');
var Promise = require('bluebird');

var request = Promise.promisifyAll(request);

var BASE_URL = 'http://localhost:8090';

describe('test the server', function() {
  var _server
  var u2f;
  var ldap_client = {
    bind: sinon.stub()
  };

  beforeEach(function(done) {
    var config = {
      port: 8090,
      totp_secret: 'totp_secret',
      ldap_url: 'ldap://127.0.0.1:389',
      ldap_users_dn: 'ou=users,dc=example,dc=com',
      session_secret: 'session_secret',
      session_max_age: 50000
    };

    u2f = {};
    u2f.startRegistration = sinon.stub();
    u2f.finishRegistration = sinon.stub();
    u2f.startAuthentication = sinon.stub();
    u2f.finishAuthentication = sinon.stub();

    ldap_client.bind.withArgs('cn=test_ok,ou=users,dc=example,dc=com', 
                              'password').yields(undefined);
    ldap_client.bind.withArgs('cn=test_nok,ou=users,dc=example,dc=com', 
                              'password').yields('error');
    _server = server.run(config, ldap_client, u2f, function() {
      done();
    });
  });

  afterEach(function() {
    _server.close();
  });

  describe('test GET /login', function() {
    test_login()
  });

  describe('test GET /logout', function() {
    test_logout()
  });

  describe('test authentication and verification', function() {
    test_authentication();
  });

  function test_login() {
    it('should serve the login page', function(done) {
      request.getAsync(BASE_URL + '/login')
      .then(function(response) {
        assert.equal(response.statusCode, 200);
        done();
      });
    });
  }
  
  function test_logout() {
    it('should logout and redirect to /', function(done) {
      request.getAsync(BASE_URL + '/logout')
      .then(function(response) {
        assert.equal(response.req.path, '/');
        done();
      });
    });
  }
  
  function test_authentication() {
    it('should return status code 401 when user is not authenticated', function() {
      return request.getAsync({ url: BASE_URL + '/verify' })
      .then(function(response) {
        assert.equal(response.statusCode, 401);
        return Promise.resolve();
      });
    });
  
    it('should return status code 204 when user is authenticated using totp', function() {
      var real_token = speakeasy.totp({
        secret: 'totp_secret',
        encoding: 'base32'
      });
      var j = request.jar();
      return request.getAsync({ url: BASE_URL + '/login', jar: j })
      .then(function(res) {
        assert.equal(res.statusCode, 200, 'get login page failed');
        return request.postAsync({ 
          url: BASE_URL + '/1stfactor',
          jar: j,
          form: {
            username: 'test_ok',
            password: 'password'
          }
        });
      }) 
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'first factor failed');
        return request.postAsync({
          url: BASE_URL + '/2ndfactor/totp',
          jar: j,
          form: {
            token: real_token
          }
        });
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'second factor failed');
        return request.getAsync({ url: BASE_URL + '/verify', jar: j })
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'verify failed');
        return Promise.resolve();
      });
    });
  
    it('should return status code 204 when user is authenticated using u2f', function() {
      var sign_request = {};
      var sign_status = {};
      var registration_request = {};
      var registration_status = {};
      u2f.startRegistration.returns(Promise.resolve(sign_request));
      u2f.finishRegistration.returns(Promise.resolve(sign_status));
      u2f.startAuthentication.returns(Promise.resolve(registration_request));
      u2f.finishAuthentication.returns(Promise.resolve(registration_status));
  
      var j = request.jar();
      return request.getAsync({ url: BASE_URL + '/login', jar: j })
      .then(function(res) {
        assert.equal(res.statusCode, 200, 'get login page failed');
        return request.postAsync({ 
          url: BASE_URL + '/1stfactor',
          jar: j,
          form: {
            username: 'test_ok',
            password: 'password'
          }
        });
      }) 
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'first factor failed');
        return request.getAsync({
          url: BASE_URL + '/2ndfactor/u2f/register_request',
          jar: j
        });
      })
      .then(function(res) {
        assert.equal(res.statusCode, 200, 'second factor, start register failed');
        return request.postAsync({
          url: BASE_URL + '/2ndfactor/u2f/register',
          jar: j,
          form: {
            s: 'test'
          }
        });
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'second factor, finish register failed');
        return request.getAsync({
          url: BASE_URL + '/2ndfactor/u2f/sign_request',
          jar: j
        });
      })
      .then(function(res) {
        assert.equal(res.statusCode, 200, 'second factor, start sign failed');
        return request.postAsync({
          url: BASE_URL + '/2ndfactor/u2f/sign',
          jar: j,
          form: {
            s: 'test'
          }
        });
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'second factor, finish sign failed');
        return request.getAsync({ url: BASE_URL + '/verify', jar: j })
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'verify failed');
        return Promise.resolve();
      });
    });
  }

});

