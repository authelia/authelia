
var request_ = require('request');
var assert = require('assert');
var speakeasy = require('speakeasy');
var j = request_.jar();
var request = request_.defaults({jar: j});
var Q = require('q');

var BASE_URL = 'http://localhost:8080';

describe('test the server', function() {
  var home_page;
  var login_page;
  var config = {
    port: 8090,
    totp_secret: 'totp_secret',
    ldap_url: 'ldap://127.0.0.1:389',
    ldap_users_dn: 'ou=users,dc=example,dc=com',
    jwt_secret: 'jwt_secret',
    jwt_expiration_time: '1h'
  };

  before(function() {
    var home_page_promise = getHomePage()
    .then(function(data) {
      home_page = data.body;
    });
    var login_page_promise = getLoginPage()
    .then(function(data) {
      login_page = data.body;
    });
    return Q.all([home_page_promise, 
                  login_page_promise]);
  });

  it('should serve the login page', function(done) {
    getPromised(BASE_URL + '/auth/login?redirect=/')
    .then(function(data) {
      assert.equal(data.response.statusCode, 200);
      done();
    });
  });

  it('should serve the homepage', function(done) {
    getPromised(BASE_URL + '/')
    .then(function(data) {
      assert.equal(data.response.statusCode, 200);
      done();
    });
  });

  it('should redirect when logout', function(done) {
    getPromised(BASE_URL + '/auth/logout?redirect=/')
    .then(function(data) {
      assert.equal(data.response.statusCode, 200);
      assert.equal(data.body, home_page);
      done();
    });
  });

  it('should be redirected to the login page when accessing secret while not authenticated', function(done) {
    getPromised(BASE_URL + '/secret.html')
    .then(function(data) {
      assert.equal(data.response.statusCode, 200);
      assert.equal(data.body, login_page);
      done();
    });
  });

  it('should fail the login', function(done) {
    postPromised(BASE_URL + '/_auth', {
      form: {
        username: 'admin',
        password: 'password',
        token: 'abc'
      }
    })
    .then(function(data) {
      assert.equal(data.body, 'Authentication failed');
      done();
    });
  });

  it('should login and access the secret', function(done) {
    var token = speakeasy.totp({
      secret: 'GRWGIJS6IRHVEODVNRCXCOBMJ5AGC6ZE',
      encoding: 'base32' 
    });
   
    postPromised(BASE_URL + '/_auth', {
      form: {
        username: 'admin',
        password: 'password',
        token: token
      }
    })
    .then(function(data) {
      assert.equal(data.response.statusCode, 200);
      assert.equal(data.body.length, 148);
      var cookie = request.cookie('access_token=' + data.body);
      j.setCookie(cookie, BASE_URL + '/_auth');
      return getPromised(BASE_URL + '/secret.html');
    })
    .then(function(data) {
      var content = data.body;
      var is_secret_page_content = 
        (content.indexOf('This is a very important secret!') > -1);
      assert(is_secret_page_content);
      done();
    })
    .fail(function(err) {
      console.error(err);
    });
  });

  it('should logoff and should not be able to access secret anymore', function(done) {
    getPromised(BASE_URL + '/secret.html')
    .then(function(data) {
      var content = data.body;
      var is_secret_page_content = 
        (content.indexOf('This is a very important secret!') > -1);
      assert(is_secret_page_content);
      return getPromised(BASE_URL + '/auth/logout')
    })
    .then(function(data) {
      assert.equal(data.response.statusCode, 200);
      assert.equal(data.body, home_page);
      return getPromised(BASE_URL + '/secret.html');
    })
    .then(function(data) {
      var content = data.body;
      assert.equal(data.body, login_page);
      done();
    })
    .fail(function(err) {
      console.error(err);
    });
  });
});

function responsePromised(defer) {
  return function(error, response, body) {
    if(error) {
      console.error(error);
      defer.reject(error);
      return;
    }
    defer.resolve({
      response: response,
      body: body
    });
  }
}

function getPromised(url) {
  var defer = Q.defer();
  request.get(url, responsePromised(defer));
  return defer.promise;
}

function postPromised(url, body) {
  var defer = Q.defer();
  request.post(url, body, responsePromised(defer));
  return defer.promise;
}

function getHomePage() {
  return getPromised(BASE_URL + '/');
}

function getLoginPage() {
  return getPromised(BASE_URL + '/auth/login');
}
