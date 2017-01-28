
var server = require('../../src/lib/server');

var Promise = require('bluebird');
var request = Promise.promisifyAll(require('request'));
var assert = require('assert');
var speakeasy = require('speakeasy');
var sinon = require('sinon');
var MockDate = require('mockdate');

var PORT = 8090;
var BASE_URL = 'http://localhost:' + PORT;
var requests = require('./requests')(PORT);

describe('test the server', function() {
  var _server
  var deps;
  var u2f, nedb;
  var transporter;
  var collection;
  var ldap_client = {
    bind: sinon.stub(),
    search: sinon.stub(),
    modify: sinon.stub(),
  };
  var ldap = {
    Change: sinon.spy()
  }

  beforeEach(function(done) {
    var config = {
      port: PORT,
      totp_secret: 'totp_secret',
      ldap_url: 'ldap://127.0.0.1:389',
      ldap_users_dn: 'ou=users,dc=example,dc=com',
      ldap_user: 'cn=admin,dc=example,dc=com',
      ldap_password: 'password',
      session_secret: 'session_secret',
      session_max_age: 50000,
      store_in_memory: true,
      gmail: {
        user: 'user@example.com',
        pass: 'password'
      }
    };

    u2f = {};
    u2f.startRegistration = sinon.stub();
    u2f.finishRegistration = sinon.stub();
    u2f.startAuthentication = sinon.stub();
    u2f.finishAuthentication = sinon.stub();

    nedb = require('nedb');
    
    transporter = {};
    transporter.sendMail = sinon.stub().yields();

    var nodemailer = {};
    nodemailer.createTransport = sinon.spy(function() {
      return transporter;
    });

    var search_doc = {
      object: {
        mail: 'test_ok@example.com'
      }
    };
 
    var search_res = {};
    search_res.on = sinon.spy(function(event, fn) {
      if(event != 'error') fn(search_doc);
    });

    ldap_client.bind.withArgs('cn=test_ok,ou=users,dc=example,dc=com', 
                              'password').yields(undefined);
    ldap_client.bind.withArgs('cn=admin,dc=example,dc=com', 
                              'password').yields(undefined);

    ldap_client.bind.withArgs('cn=test_nok,ou=users,dc=example,dc=com', 
                              'password').yields('error');

    ldap_client.modify.yields(undefined);
    ldap_client.search.yields(undefined, search_res);

    var deps = {};
    deps.u2f = u2f;
    deps.nedb = nedb;
    deps.nodemailer = nodemailer;
    deps.ldap = ldap;

    _server = server.run(config, ldap_client, deps, function() {
      done();
    });
  });

  afterEach(function() {
    _server.close();
  });

  describe('test GET /login', function() {
    test_login();
  });

  describe('test GET /logout', function() {
    test_logout();
  });

  describe('test GET /reset-password-form', function() {
    test_reset_password_form();
  });

  describe('test endpoints locks', function() {
    function should_post_and_reply_with(url, status_code) {
      return request.postAsync(url).then(function(response) {
        assert.equal(response.statusCode, status_code);
        return Promise.resolve();
      }) 
    }

    function should_get_and_reply_with(url, status_code) {
      return request.getAsync(url).then(function(response) {
        assert.equal(response.statusCode, status_code);
        return Promise.resolve();
      }) 
    }

    function should_post_and_reply_with_403(url) {
      return should_post_and_reply_with(url, 403);
    }
    function should_get_and_reply_with_403(url) {
      return should_get_and_reply_with(url, 403);
    }

    function should_post_and_reply_with_401(url) {
      return should_post_and_reply_with(url, 401);
    }
    function should_get_and_reply_with_401(url) {
      return should_get_and_reply_with(url, 401);
    }

    function should_get_and_post_reply_with_403(url) {
      var p1 = should_post_and_reply_with_403(url);
      var p2 = should_get_and_reply_with_403(url);
      return Promise.all([p1, p2]);
    }

    it('should block /authentication/new-password', function() {
      return should_post_and_reply_with_403(BASE_URL + '/authentication/new-password')
    });

    it('should block /authentication/u2f-register', function() {
      return should_get_and_post_reply_with_403(BASE_URL + '/authentication/u2f-register');
    });

    it('should block /authentication/reset-password', function() {
      return should_get_and_post_reply_with_403(BASE_URL + '/authentication/reset-password');
    });

    it('should block /authentication/2ndfactor/u2f/register_request', function() {
      return should_get_and_reply_with_403(BASE_URL + '/authentication/2ndfactor/u2f/register_request');
    });

    it('should block /authentication/2ndfactor/u2f/register', function() {
      return should_post_and_reply_with_403(BASE_URL + '/authentication/2ndfactor/u2f/register');
    });

    it('should block /authentication/2ndfactor/u2f/sign_request', function() {
      return should_get_and_reply_with_403(BASE_URL + '/authentication/2ndfactor/u2f/sign_request');
    });

    it('should block /authentication/2ndfactor/u2f/sign', function() {
      return should_post_and_reply_with_403(BASE_URL + '/authentication/2ndfactor/u2f/sign');
    });
  });

  describe('test authentication and verification', function() {
    test_authentication();
    test_reset_password();
    test_regulation();
  });

  function test_reset_password_form() {
    it('should serve the reset password form page', function(done) {
      request.getAsync(BASE_URL + '/authentication/reset-password-form')
      .then(function(response) {
        assert.equal(response.statusCode, 200);
        done();
      });
    });
  }

  function test_login() {
    it('should serve the login page', function(done) {
      request.getAsync(BASE_URL + '/authentication/login')
      .then(function(response) {
        assert.equal(response.statusCode, 200);
        done();
      });
    });
  }
  
  function test_logout() {
    it('should logout and redirect to /', function(done) {
      request.getAsync(BASE_URL + '/authentication/logout')
      .then(function(response) {
        assert.equal(response.req.path, '/');
        done();
      });
    });
  }
  
  function test_authentication() {
    it('should return status code 401 when user is not authenticated', function() {
      return request.getAsync({ url: BASE_URL + '/authentication/verify' })
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
      return requests.login(j)
      .then(function(res) {
        assert.equal(res.statusCode, 200, 'get login page failed');
        return requests.first_factor(j);
      }) 
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'first factor failed');
        return requests.totp(j, real_token);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'second factor failed');
        return requests.verify(j);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'verify failed');
        return Promise.resolve();
      });
    });
  
    it('should keep session variables when login page is reloaded', function() {
      var real_token = speakeasy.totp({
        secret: 'totp_secret',
        encoding: 'base32'
      });
      var j = request.jar();
      return requests.login(j)
      .then(function(res) {
        assert.equal(res.statusCode, 200, 'get login page failed');
        return requests.first_factor(j);
      }) 
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'first factor failed');
        return requests.totp(j, real_token);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'second factor failed');
        return requests.login(j);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 200, 'login page loading failed');
        return requests.verify(j);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'verify failed');
        return Promise.resolve();
      })
      .catch(function(err) {
        console.error(err);
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
      return requests.login(j)
      .then(function(res) {
        assert.equal(res.statusCode, 200, 'get login page failed');
        return requests.first_factor(j);
      }) 
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'first factor failed');
        return requests.u2f_registration(j, transporter);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'second factor, finish register failed');
        return requests.u2f_authentication(j);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'second factor, finish sign failed');
        return requests.verify(j);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'verify failed');
        return Promise.resolve();
      });
    });
  }
 
  function test_reset_password() {
    it('should reset the password', function() {
      var j = request.jar();
      return requests.login(j)
      .then(function(res) {
        assert.equal(res.statusCode, 200, 'get login page failed');
        return requests.first_factor(j);
      }) 
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'first factor failed');
        return requests.reset_password(j, transporter, 'user', 'new-password');
      })
      .then(function(res) {
        assert.equal(res.statusCode, 204, 'second factor, finish register failed');
        return Promise.resolve();
      });
    });
  }

  function test_regulation() {
    it('should regulate authentication', function() {
      var j = request.jar();
      MockDate.set('1/2/2017 00:00:00');
      return requests.login(j)
      .then(function(res) {
        assert.equal(res.statusCode, 200, 'get login page failed');
        return requests.failing_first_factor(j);
      }) 
      .then(function(res) {
        console.log('coucou');
        assert.equal(res.statusCode, 401, 'first factor failed');
        return requests.failing_first_factor(j);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 401, 'first factor failed');
        return requests.failing_first_factor(j);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 401, 'first factor failed');
        return requests.failing_first_factor(j);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 403, 'first factor failed');
        MockDate.set('1/2/2017 00:30:00');
        return requests.failing_first_factor(j);
      })
      .then(function(res) {
        assert.equal(res.statusCode, 401, 'first factor failed');
        return Promise.resolve();
      })
    });
  }
});

