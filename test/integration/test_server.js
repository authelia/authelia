
var request_ = require('request');
var assert = require('assert');
var speakeasy = require('speakeasy');
var j = request_.jar();
var Promise = require('bluebird');
var request = Promise.promisifyAll(request_.defaults({jar: j}));

var BASE_URL = 'https://localhost:8080';

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

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
    getPromised(BASE_URL + '/auth/login?redirect=/')
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
    getPromised(BASE_URL + '/auth/logout?redirect=/')
    .then(function(data) {
      assert.equal(data.statusCode, 200);
      assert.equal(data.body, home_page);
      done();
    });
  });

  it('should be redirected to the login page when accessing secret while not authenticated', function() {
    return getPromised(BASE_URL + '/secret.html')
    .then(function(data) {
      assert.equal(data.statusCode, 200);
      assert.equal(data.body, login_page);
      return Promise.resolve();
    });
  });

  it('should fail the first_factor login', function() {
    return postPromised(BASE_URL + '/auth/1stfactor', {
      form: {
        username: 'admin',
        password: 'bad_password'
      }
    })
    .then(function(data) {
      assert.equal(401, data.statusCode);
      return Promise.resolve();
    });
  });

  it('should login and access the secret using totp', function() {
    var token = speakeasy.totp({
      secret: 'GRWGIJS6IRHVEODVNRCXCOBMJ5AGC6ZE',
      encoding: 'base32' 
    });
   
    return postPromised(BASE_URL + '/auth/1stfactor', {
      form: {
        username: 'admin',
        password: 'password',
      }
    })
    .then(function(response) {
      assert.equal(response.statusCode, 204);
      return postPromised(BASE_URL + '/auth/2ndfactor/totp', {
        form: { token: token }
      });
    })
    .then(function(response) {
      assert.equal(response.statusCode, 204);
      return getPromised(BASE_URL + '/secret.html');
    })
    .then(function(response) {
      var content = response.body;
      var is_secret_page_content = 
        (content.indexOf('This is a very important secret!') > -1);
      assert(is_secret_page_content);
      return Promise.resolve();
    })
    .catch(function(err) {
      console.error(err);
      return Promise.reject(err);
    });
  });

  it('should logoff and should not be able to access secret anymore', function() {
    return getPromised(BASE_URL + '/secret.html')
    .then(function(data) {
      var content = data.body;
      var is_secret_page_content = 
        (content.indexOf('This is a very important secret!') > -1);
      assert(is_secret_page_content);
      return getPromised(BASE_URL + '/auth/logout')
    })
    .then(function(data) {
      assert.equal(data.statusCode, 200);
      assert.equal(data.body, home_page);
      return getPromised(BASE_URL + '/secret.html');
    })
    .then(function(data) {
      var content = data.body;
      assert.equal(data.body, login_page);
      return Promise.resolve();
    })
    .catch(function(err) {
      console.error(err);
      return Promise.reject();
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
  console.log('GET: %s', url);
  return request.getAsync(url);
}

function postPromised(url, body) {
  console.log('POST: %s, %s', url, JSON.stringify(body));
  return request.postAsync(url, body);
}

function getHomePage() {
  return getPromised(BASE_URL + '/');
}

function getLoginPage() {
  return getPromised(BASE_URL + '/auth/login');
}
