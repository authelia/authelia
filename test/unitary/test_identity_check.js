
var sinon = require('sinon');
var identity_check = require('../../src/lib/identity_check');
var exceptions = require('../../src/lib/exceptions');
var assert = require('assert');
var winston = require('winston');
var Promise = require('bluebird');

describe('test identity check process', function() {
  var req, res, app, icheck_interface;
  var user_data_store;
  var email_sender;

  beforeEach(function() {
    req = {};
    res = {};

    app = {};
    icheck_interface = {};
    icheck_interface.pre_check_callback = sinon.stub();

    user_data_store = {};
    user_data_store.issue_identity_check_token = sinon.stub();
    user_data_store.issue_identity_check_token.returns(Promise.resolve());
    user_data_store.consume_identity_check_token = sinon.stub();
    user_data_store.consume_identity_check_token.returns(Promise.resolve({ userid: 'user' }));

    email_sender = {};
    email_sender.send = sinon.stub();
    email_sender.send = sinon.stub().returns(Promise.resolve());

    req.headers = {};
    req.session = {};
    req.session.auth_session = {};

    req.query = {};
    req.app = {};
    req.app.get = sinon.stub();
    req.app.get.withArgs('logger').returns(winston);
    req.app.get.withArgs('user data store').returns(user_data_store);
    req.app.get.withArgs('email sender').returns(email_sender);

    res.status = sinon.spy();
    res.send = sinon.spy();
    res.redirect = sinon.spy();
    res.render = sinon.spy();

    app.get = sinon.spy();
    app.post = sinon.spy();
  });

  it('should register a POST and GET endpoint', function() {
    var app = {};
    app.get = sinon.spy();
    app.post = sinon.spy();
    var endpoint = '/test';
    var icheck_interface = {};

    identity_check(app, endpoint, icheck_interface);

    assert(app.get.calledOnce);
    assert(app.get.calledWith(endpoint));

    assert(app.post.calledOnce);
    assert(app.post.calledWith(endpoint));
  });

  describe('test POST', test_post_handler);
  describe('test GET', test_get_handler);

  function test_post_handler() {
    it('should send 403 if pre check rejects', function(done) {
      var endpoint = '/protected';

      icheck_interface.pre_check_callback.returns(Promise.reject('No access'));
      identity_check(app, endpoint, icheck_interface);

      res.send = sinon.spy(function() {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });

      var handler = app.post.getCall(0).args[1];
      handler(req, res);
    });

    it('should send 400 if email is missing in provided identity', function(done) {
      var endpoint = '/protected';
      var identity = { userid: 'abc' };

      icheck_interface.pre_check_callback.returns(Promise.resolve(identity));
      identity_check(app, endpoint, icheck_interface);

      res.send = sinon.spy(function() {
        assert.equal(res.status.getCall(0).args[0], 400);
        done();
      });

      var handler = app.post.getCall(0).args[1];
      handler(req, res);
    });

    it('should send 400 if userid is missing in provided identity', function(done) {
      var endpoint = '/protected';
      var identity = { email: 'abc@example.com' };

      icheck_interface.pre_check_callback.returns(Promise.resolve(identity));
      identity_check(app, endpoint, icheck_interface);

      res.send = sinon.spy(function() {
        assert.equal(res.status.getCall(0).args[0], 400);
        done();
      });
      var handler = app.post.getCall(0).args[1];
      handler(req, res);
    });

    it('should issue a token, send an email and return 204', function(done) {
      var endpoint = '/protected';
      var identity = { userid: 'user', email: 'abc@example.com' };
      req.headers.host = 'localhost';
      req.headers['x-original-uri'] = '/auth/test';

      icheck_interface.pre_check_callback.returns(Promise.resolve(identity));
      identity_check(app, endpoint, icheck_interface);

      res.send = sinon.spy(function() {
        assert.equal(res.status.getCall(0).args[0], 204);
        assert(email_sender.send.calledOnce);
        assert(user_data_store.issue_identity_check_token.calledOnce);
        assert.equal(user_data_store.issue_identity_check_token.getCall(0).args[0], 'user');
        assert.equal(user_data_store.issue_identity_check_token.getCall(0).args[3], 240000);
        done();
      });
      var handler = app.post.getCall(0).args[1];
      handler(req, res);
    });
  }

  function test_get_handler() {
    it('should send 403 if no identity_token is provided', function(done) {
      var endpoint = '/protected';

      identity_check(app, endpoint, icheck_interface);

      res.send = sinon.spy(function() {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });
      var handler = app.get.getCall(0).args[1];
      handler(req, res);
    });

    it('should render template if identity_token is provided and still valid', function(done) {
      req.query.identity_token = 'token';
      var endpoint = '/protected';

      icheck_interface.render_template = 'template';

      identity_check(app, endpoint, icheck_interface);

      res.render = sinon.spy(function(template) {
        assert.equal(template, 'template');
        done();
      });
      var handler = app.get.getCall(0).args[1];
      handler(req, res);
    });

    it('should return 403 if identity_token is provided but invalid', function(done) {
      req.query.identity_token = 'token';
      var endpoint = '/protected';

      icheck_interface.render_template = 'template';
      user_data_store.consume_identity_check_token
        .returns(Promise.reject('Invalid token'));

      identity_check(app, endpoint, icheck_interface);

      res.send = sinon.spy(function(template) {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });
      var handler = app.get.getCall(0).args[1];
      handler(req, res);
    });

    it('should set the identity_check session object even if session does not exist yet', function(done) {
      req.query.identity_token = 'token';
      var endpoint = '/protected';

      req.session = {};
      icheck_interface.render_template = 'template';

      identity_check(app, endpoint, icheck_interface);

      res.render = sinon.spy(function(template) {
        assert.equal(req.session.auth_session.identity_check.userid, 'user');
        assert.equal(template, 'template');
        done();
      });
      var handler = app.get.getCall(0).args[1];
      handler(req, res);
    });
  }
});
