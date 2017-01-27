
var server = require('../../src/lib/server');

var Promise = require('bluebird');
var request = Promise.promisifyAll(require('request'));
var assert = require('assert');
var speakeasy = require('speakeasy');
var sinon = require('sinon');
var tmp = require('tmp');
var nedb = require('nedb');

var PORT = 8050;
var BASE_URL = 'http://localhost:' + PORT;

var requests = require('./requests')(PORT);


describe('test data persistence', function() {
  var u2f;
  var tmpDir;
  var ldap_client = {
    bind: sinon.stub(),
    search: sinon.stub()
  };
  var config;

  before(function() {
    u2f = {};
    u2f.startRegistration = sinon.stub();
    u2f.finishRegistration = sinon.stub();
    u2f.startAuthentication = sinon.stub();
    u2f.finishAuthentication = sinon.stub();

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
    ldap_client.bind.withArgs('cn=test_nok,ou=users,dc=example,dc=com', 
                              'password').yields('error');
    ldap_client.search.yields(undefined, search_res);

    tmpDir = tmp.dirSync({ unsafeCleanup: true });
    config = {
      port: PORT,
      totp_secret: 'totp_secret',
      ldap_url: 'ldap://127.0.0.1:389',
      ldap_users_dn: 'ou=users,dc=example,dc=com',
      session_secret: 'session_secret',
      session_max_age: 50000,
      store_directory: tmpDir.name,
      gmail: { user: 'user@example.com', pass: 'password' }
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

    var nodemailer = {};
    var transporter = {
      sendMail: sinon.stub().yields()
    };
    nodemailer.createTransport = sinon.spy(function() {
      return transporter;
    });

    var deps = {};
    deps.u2f = u2f;
    deps.nedb = nedb;
    deps.nodemailer = nodemailer;

    var j1 = request.jar();
    var j2 = request.jar();

    return start_server(config, ldap_client, deps)
    .then(function(s) {
      server = s;
      return requests.login(j1);
    })
    .then(function(res) {
      return requests.first_factor(j1);
    }) 
    .then(function() {
      return requests.u2f_registration(j1, transporter);
    })
    .then(function() {
      return requests.u2f_authentication(j1);
    })
    .then(function() {
      return stop_server(server);
    })
    .then(function() {
      return start_server(config, ldap_client, deps)
    })
    .then(function(s) {
      server = s;
      return requests.login(j2);
    })
    .then(function() {
      return requests.first_factor(j2);
    }) 
    .then(function() {
      return requests.u2f_authentication(j2);
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

  function start_server(config, ldap_client, deps) {
    return new Promise(function(resolve, reject) {
      var s = server.run(config, ldap_client, deps);
      resolve(s);
    });
  }

  function stop_server(s) {
    return new Promise(function(resolve, reject) {
      s.close();
      resolve();
    });
  }
});
