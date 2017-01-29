
var sinon = require('sinon');
var Promise = require('bluebird');
var assert = require('assert');
var u2f = require('../../../src/lib/routes/u2f');
var winston = require('winston');

describe('test u2f routes', function() {
  var req, res;
  var user_data_store;

  beforeEach(function() {
    req = {}
    req.app = {};
    req.app.get = sinon.stub();
    req.app.get.withArgs('logger').returns(winston);
    req.session = {};
    req.session.auth_session = {};
    req.session.auth_session.userid = 'user';
    req.session.auth_session.first_factor = true;
    req.session.auth_session.second_factor = false;
    req.session.auth_session.identity_check = {};
    req.session.auth_session.identity_check.challenge = 'u2f-register';
    req.session.auth_session.register_request = {};
    req.headers = {};
    req.headers.host = 'localhost';

    var options = {};
    options.inMemoryOnly = true;

    user_data_store = {};
    user_data_store.set_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    user_data_store.get_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    req.app.get.withArgs('user data store').returns(user_data_store);

    res = {};
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  })

  describe('test registration request', test_registration_request);
  describe('test registration', test_registration);
  describe('test signing request', test_signing_request);
  describe('test signing', test_signing);

  function test_registration_request() {
    it('should send back the registration request and save it in the session', function(done) {
      var expectedRequest = {
        test: 'abc'
      };
      res.json = sinon.spy(function(data) {
        assert.equal(200, res.status.getCall(0).args[0]);
        assert.deepEqual(expectedRequest, data);
        done();
      });
      var user_key_container = {};
      var u2f_mock = {};
      u2f_mock.startRegistration = sinon.stub();
      u2f_mock.startRegistration.returns(Promise.resolve(expectedRequest));

      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.register_request(req, res);
    });

    it('should return internal error on registration request', function(done) {
      res.send = sinon.spy(function(data) {
        assert.equal(500, res.status.getCall(0).args[0]);
        done();
      });
      var user_key_container = {};
      var u2f_mock = {};
      u2f_mock.startRegistration = sinon.stub();
      u2f_mock.startRegistration.returns(Promise.reject('Internal error'));

      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.register_request(req, res);
    });

    it('should return forbidden if identity has not been verified', function(done) {
      res.send = sinon.spy(function(data) {
        assert.equal(403, res.status.getCall(0).args[0]);
        done();
      });
      req.session.auth_session.identity_check = undefined;
      u2f.register_request(req, res);
    });
  }

  function test_registration() {
    it('should save u2f meta and return status code 200', function(done) {
      var expectedStatus = {
        keyHandle: 'keyHandle',
        publicKey: 'pbk',
        certificate: 'cert'
      };
      res.send = sinon.spy(function(data) {
        assert.equal('user', user_data_store.set_u2f_meta.getCall(0).args[0])
        assert.equal(req.session.auth_session.identity_check, undefined);
        done();
      });
      var u2f_mock = {};
      u2f_mock.finishRegistration = sinon.stub();
      u2f_mock.finishRegistration.returns(Promise.resolve(expectedStatus));

      req.session.auth_session.register_request = {};
      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.register(req, res);
    });

    it('should return unauthorized on finishRegistration error', function(done) {
      res.send = sinon.spy(function(data) {
        assert.equal(500, res.status.getCall(0).args[0]);
        done();
      });
      var user_key_container = {};
      var u2f_mock = {};
      u2f_mock.finishRegistration = sinon.stub();
      u2f_mock.finishRegistration.returns(Promise.reject('Internal error'));

      req.session.auth_session.register_request = 'abc';
      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.register(req, res);
    });

    it('should return 403 when register_request is not provided', function(done) {
      res.send = sinon.spy(function(data) {
        assert.equal(403, res.status.getCall(0).args[0]);
        done();
      });
      var user_key_container = {};
      var u2f_mock = {};
      u2f_mock.finishRegistration = sinon.stub();
      u2f_mock.finishRegistration.returns(Promise.resolve());

      req.session.auth_session.register_request = undefined;
      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.register(req, res);
    });

    it('should return forbidden error when no auth request has been initiated', function(done) {
      res.send = sinon.spy(function(data) {
        assert.equal(403, res.status.getCall(0).args[0]);
        done();
      });
      var user_key_container = {};
      var u2f_mock = {};
      u2f_mock.finishRegistration = sinon.stub();
      u2f_mock.finishRegistration.returns(Promise.resolve());

      req.session.auth_session.register_request = undefined;
      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.register(req, res);
    });

    it('should return forbidden error when identity has not been verified', function(done) {
      res.send = sinon.spy(function(data) {
        assert.equal(403, res.status.getCall(0).args[0]);
        done();
      });
      req.session.auth_session.identity_check = undefined;
      u2f.register(req, res);
    });
  }
  
  function test_signing_request() {
    it('should send back the sign request and save it in the session', function(done) {
      var expectedRequest = {
        test: 'abc'
      };
      res.json = sinon.spy(function(data) {
        assert.deepEqual(expectedRequest, req.session.auth_session.sign_request);
        assert.equal(200, res.status.getCall(0).args[0]);
        assert.deepEqual(expectedRequest, data);
        done();
      });
      var user_key_container = {};
      user_key_container['user'] = {}; // simulate a registration
      var u2f_mock = {};
      u2f_mock.startAuthentication = sinon.stub();
      u2f_mock.startAuthentication.returns(Promise.resolve(expectedRequest));

      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.sign_request(req, res);
    });

    it('should return unauthorized error on registration request error', function(done) {
      res.send = sinon.spy(function(data) {
        assert.equal(500, res.status.getCall(0).args[0]);
        done();
      });
      var user_key_container = {};
      user_key_container['user'] = {}; // simulate a registration
      var u2f_mock = {};
      u2f_mock.startAuthentication = sinon.stub();
      u2f_mock.startAuthentication.returns(Promise.reject('Internal error'));

      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.sign_request(req, res);
    });

    it('should send unauthorized error when no registration exists', function(done) {
      var expectedRequest = {
        test: 'abc'
      };
      res.send = sinon.spy(function(data) {
        assert.equal(401, res.status.getCall(0).args[0]);
        done();
      });
      var user_key_container = {}; // no entry means no registration
      var u2f_mock = {};
      u2f_mock.startAuthentication = sinon.stub();
      u2f_mock.startAuthentication.returns(Promise.resolve(expectedRequest));

      user_data_store.get_u2f_meta = sinon.stub().returns(Promise.resolve());

      req.app.get = sinon.stub();
      req.app.get.withArgs('logger').returns(winston);
      req.app.get.withArgs('user data store').returns(user_data_store);
      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.sign_request(req, res);
    });
  }

  function test_signing() {
    it('should return status code 204', function(done) {
      var user_key_container = {};
      user_key_container['user'] = {};
      var expectedStatus = {
        keyHandle: 'keyHandle',
        publicKey: 'pbk',
        certificate: 'cert'
      };
      res.send = sinon.spy(function(data) {
        assert(204, res.status.getCall(0).args[0]);
        assert(req.session.auth_session.second_factor);
        done();
      });
      var u2f_mock = {};
      u2f_mock.finishAuthentication = sinon.stub();
      u2f_mock.finishAuthentication.returns(Promise.resolve(expectedStatus));

      req.session.auth_session.sign_request = {};
      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.sign(req, res);
    });

    it('should return unauthorized error on registration request internal error', function(done) {
      res.send = sinon.spy(function(data) {
        assert.equal(500, res.status.getCall(0).args[0]);
        done();
      });
      var user_key_container = {};
      user_key_container['user'] = {};

      var u2f_mock = {};
      u2f_mock.finishAuthentication = sinon.stub();
      u2f_mock.finishAuthentication.returns(Promise.reject('Internal error'));

      req.session.auth_session.sign_request = {};
      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.sign(req, res);
    });

    it('should return unauthorized error when no sign request has been initiated', function(done) {
      res.send = sinon.spy(function(data) {
        assert.equal(401, res.status.getCall(0).args[0]);
        done();
      });
      var user_key_container = {};
      var u2f_mock = {};
      u2f_mock.finishAuthentication = sinon.stub();
      u2f_mock.finishAuthentication.returns(Promise.resolve());

      req.app.get.withArgs('u2f').returns(u2f_mock);
      u2f.sign(req, res);
    });
  }
});

