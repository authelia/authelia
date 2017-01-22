
var sinon = require('sinon');
var winston = require('winston');
var u2f_register = require('../../../src/lib/routes/u2f_register');
var assert = require('assert');

describe('test register handle', function() {
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
    req.session.auth_session.email = 'user@example.com';
    req.session.auth_session.first_factor = true;
    req.session.auth_session.second_factor = false;
    req.headers = {};
    req.headers.host = 'localhost';

    var options = {};
    options.inMemoryOnly = true;

    user_data_store = {};
    user_data_store.set_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    user_data_store.get_u2f_meta = sinon.stub().returns(Promise.resolve({}));
    user_data_store.save_u2f_registration_token = sinon.stub().returns(Promise.resolve({}));
    user_data_store.verify_u2f_registration_token = sinon.stub().returns(Promise.resolve({}));
    req.app.get.withArgs('user data store').returns(user_data_store);

    res = {};
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  })


  describe('test registration handler (POST)', test_registration_handler_post);
  describe('test registration handler (GET)', test_registration_handler_get);

  function test_registration_handler_post() {
    it('should issue a registration token', function(done) {
      res.send = sinon.spy(function() {
        assert.equal(204, res.status.getCall(0).args[0]);
        assert.equal('user', user_data_store.save_u2f_registration_token.getCall(0).args[0]);
        assert.equal(4 * 60 * 1000, user_data_store.save_u2f_registration_token.getCall(0).args[2]);
        done();
      });
      var email_sender = {};
      email_sender.send = sinon.stub().returns(Promise.resolve());
      req.app.get.withArgs('email sender').returns(email_sender);
      u2f_register.register_handler_post(req, res);
    });

    it('should fail during issuance of a registration token', function(done) {
      res.send = sinon.spy(function() {
        assert.equal(500, res.status.getCall(0).args[0]);
        done();
      });
      user_data_store.save_u2f_registration_token = sinon.stub().returns(Promise.reject('Error'));
      u2f_register.register_handler_post(req, res);
    });

    it('should send bad request if no email has been found for the given user', function(done) {
      res.send = sinon.spy(function() {
        assert.equal(400, res.status.getCall(0).args[0]);
        done();
      });
      req.session.auth_session.email = undefined;
      var email_sender = {};
      email_sender.send = sinon.stub().returns(Promise.resolve());
      req.app.get.withArgs('email sender').returns(email_sender);

      u2f_register.register_handler_post(req, res);
    });
  }

  function test_registration_handler_get() {
    it('should send forbidden if no registration_token has been provided', function(done) {
      res.send = sinon.spy(function() {
        assert.equal(403, res.status.getCall(0).args[0]);
        done();
      });
      u2f_register.register_handler_get(req, res);
    });


    it('should render the u2f-register view when registration token is still valid', function(done) {
      res.render = sinon.spy(function(data) {
        assert.equal('u2f_register', data);
        done();
      });
      req.query = {};
      req.query.registration_token = 'token';
      u2f_register.register_handler_get(req, res);
    });

    it('should send forbidden status when registration token is not valid', function(done) {
      res.send = sinon.spy(function(data) {
        assert.equal(403, res.status.getCall(0).args[0]);
        done();
      });

      req.params = {};
      req.params.registration_token = 'token';
      user_data_store.verify_u2f_registration_token = sinon.stub().returns(Promise.reject('Not valid anymore'));

      u2f_register.register_handler_get(req, res);
    });
  }
});
