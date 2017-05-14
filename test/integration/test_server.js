
var request_ = require('request');
var assert = require('assert');
var speakeasy = require('speakeasy');
var j = request_.jar();
var Promise = require('bluebird');
var request = Promise.promisifyAll(request_.defaults({jar: j}));
var util = require('util');
var sinon = require('sinon');

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

var AUTHELIA_HOST = 'nginx';
var DOMAIN = 'test.local';
var PORT = 8080;

var BASE_URL = util.format('https://%s.%s:%d', 'home', DOMAIN, PORT);
var BASE_AUTH_URL = util.format('https://%s.%s:%d/authentication', 'auth', DOMAIN, PORT);

describe('test the server', function() {
  var home_page;
  var login_page;

  before(function() {
    var home_page_promise = getHomePage()
    .then(function(data) {
      home_page = data.body;
    });
    var login_page_promise = getLoginPage()
    .then(function(data) {
      login_page = data.body;
    });
    return Promise.all([home_page_promise, 
                        login_page_promise]);
  });

  it('should serve the login page', function(done) {
    getPromised(BASE_AUTH_URL + '/login?redirect=/')
    .then(function(data) {
      assert.equal(data.statusCode, 200);
      done();
    });
  });

  it('should serve the homepage', function(done) {
    getPromised(BASE_URL + '/')
    .then(function(data) {
      assert.equal(data.statusCode, 200);
      done();
    });
  });

  it('should redirect when logout', function(done) {
    getPromised(BASE_AUTH_URL + '/logout?redirect=' + BASE_URL)
    .then(function(data) {
      assert.equal(data.statusCode, 200);
      assert.equal(data.body, home_page);
      done();
    });
  });

  it('should be redirected to the login page when accessing secret while not authenticated', function(done) {
    var url = BASE_URL + '/secret.html';
    // console.log(url);
    getPromised(url)
    .then(function(data) {
      assert.equal(data.statusCode, 200);
      assert.equal(data.body, login_page);
      done();
    });
  });

  it.skip('should fail the first factor', function(done) {
    postPromised(BASE_AUTH_URL + '/1stfactor', {
      form: {
        username: 'admin',
        password: 'password',
      }
    })
    .then(function(data) {
      assert.equal(data.body, 'Bad credentials');
      done();
    });
  });

  function login_as(username, password) {
    return postPromised(BASE_AUTH_URL + '/1stfactor', {
      form: {
        username: 'john',
        password: 'password',
      }
    })
    .then(function(data) {
      assert.equal(data.statusCode, 204);
      return Promise.resolve();
    });
  }

  it('should succeed the first factor', function() {
    return login_as('john', 'password');
  });

  describe('test ldap connection', function() {
    it('should not fail after inactivity', function() {
      var clock = sinon.useFakeTimers(); 
      return login_as('john', 'password')
      .then(function() {
        clock.tick(3600000 * 24); // 24 hour
        return login_as('john', 'password');
      })
      .then(function() {
        clock.restore();
        return Promise.resolve();
      });
    });
  });
});

function getPromised(url) {
  return request.getAsync(url);
}

function postPromised(url, body) {
  return request.postAsync(url, body);
}

function getHomePage() {
  return getPromised(BASE_URL + '/');
}

function getLoginPage() {
  return getPromised(BASE_AUTH_URL + '/login');
}
