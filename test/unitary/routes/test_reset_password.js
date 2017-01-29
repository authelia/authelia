var sinon = require('sinon');
var winston = require('winston');
var reset_password = require('../../../src/lib/routes/reset_password');
var assert = require('assert');

describe('test reset password', function() {
  var req, res;
  var user_data_store;
  var ldap_client;
  var ldap;

  beforeEach(function() {
    req = {}
    req.body = {};
    req.body.userid = 'user';
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
    user_data_store.issue_identity_check_token = sinon.stub().returns(Promise.resolve({}));
    user_data_store.consume_identity_check_token = sinon.stub().returns(Promise.resolve({}));
    req.app.get.withArgs('user data store').returns(user_data_store);

    ldap = {};
    ldap.Change = sinon.spy();
    req.app.get.withArgs('ldap').returns(ldap);

    ldap_client = {};
    ldap_client.bind = sinon.stub();
    ldap_client.search = sinon.stub();
    ldap_client.modify = sinon.stub();
    req.app.get.withArgs('ldap client').returns(ldap_client);

    config = {};
    config.ldap_users_dn = 'dc=example,dc=com';
    req.app.get.withArgs('config').returns(config);

    res = {};
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  });

  describe('test reset password identity pre check', test_reset_password_check);
  describe('test reset password post', test_reset_password_post);

  function test_reset_password_check() {
    it('should fail when no userid is provided', function(done) {
      req.body.userid = undefined;
      reset_password.icheck_interface.pre_check_callback(req)
      .catch(function(err) {
        done();
      });
    });

    it('should fail if ldap fail', function(done) {
      ldap_client.search.yields('Internal error'); 
      reset_password.icheck_interface.pre_check_callback(req)
      .catch(function(err) {
        done();
      });
    });

    it('should returns identity when ldap replies', function(done) {
      var doc = {};
      doc.object = {};
      doc.object.email = 'test@example.com';
      doc.object.userid = 'user';

      var res = {};
      res.on = sinon.stub();
      res.on.withArgs('searchEntry').yields(doc);
      res.on.withArgs('end').yields();

      ldap_client.search.yields(undefined, res); 
      reset_password.icheck_interface.pre_check_callback(req)
      .then(function() {
        done();
      });
    });
  }

  function test_reset_password_post() {
    it('should update the password and reset auth_session for reauthentication', function(done) {
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.userid = 'user';
      req.session.auth_session.identity_check.challenge = 'reset-password';
      req.body = {};
      req.body.password = 'new-password';

      ldap_client.modify.yields(undefined);
      ldap_client.bind.yields(undefined);
      res.send = sinon.spy(function() {
        assert.equal(ldap_client.modify.getCall(0).args[0], 'cn=user,dc=example,dc=com');
        assert.equal(res.status.getCall(0).args[0], 204);
        assert.equal(req.session.auth_session, undefined);
        done();
      });
      reset_password.post(req, res); 
    });

    it('should fail if identity_challenge does not exist', function(done) {
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.challenge = undefined;
      res.send = sinon.spy(function() {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });
      reset_password.post(req, res); 
    });

    it('should fail when ldap fails', function(done) {
      req.session.auth_session.identity_check = {};
      req.session.auth_session.identity_check.challenge = 'reset-password';
      req.body = {};
      req.body.password = 'new-password';

      ldap_client.bind.yields(undefined);
      ldap_client.modify.yields('Internal error with LDAP');
      res.send = sinon.spy(function() {
        assert.equal(res.status.getCall(0).args[0], 500);
        done();
      });
      reset_password.post(req, res); 
    });
  }
});
