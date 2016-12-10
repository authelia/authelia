
var assert = require('assert');
var authentication = require('../lib/authentication');
var create_res_mock = require('./res_mock');
var sinon = require('sinon');
var sinonPromise = require('sinon-promise');
sinonPromise(sinon);

var autoResolving = sinon.promise().resolves();

function create_req_mock(token) {
  return {
    body: {
      username: 'username',
      password: 'password',
      token: token
    },
    cookies: {
      'access_token': 'cookie_token'
    }
  }
}

function create_mocks() {
  var totp_token = 'totp_token';
  var jwt_token = 'jwt_token';

  var res_mock = create_res_mock();
  var req_mock = create_req_mock(totp_token);
  var bind_mock = sinon.mock();
  var totp_mock = sinon.mock();
  var sign_mock = sinon.mock();
  var verify_mock = sinon.promise();
  var jwt = {
    sign: sign_mock,
    verify: verify_mock
  };
  var ldap_interface_mock = {
    bind: bind_mock
  };
  var totp_interface_mock = {
    totp: totp_mock
  };

  bind_mock.yields();
  totp_mock.returns(totp_token);
  sign_mock.returns(jwt_token);

  var args = {
    totp_secret: 'totp_secret',
    jwt: jwt,
    jwt_expiration_time: '1h',
    users_dn: 'dc=example,dc=com',
    ldap_interface: ldap_interface_mock,
    totp_interface: totp_interface_mock
  }

  return {
    req: req_mock, 
    res: res_mock, 
    args: args,
    totp: totp_mock,
    jwt: jwt
  }
}

describe('test jwt', function() {
  describe('test authentication', function() {
    it('should authenticate user successfuly', function(done) {
      var jwt_token = 'jwt_token';
      var clock = sinon.useFakeTimers();
      var mocks = create_mocks();
      authentication.authenticate(mocks.req, mocks.res, mocks.args)
      .then(function() {
        clock.restore();
        assert(mocks.res.cookie.calledWith('access_token', jwt_token));
        assert(mocks.res.redirect.calledWith('/'));
        done();
      })
    });

    it('should fail authentication', function(done) {
      var clock = sinon.useFakeTimers();
      var mocks = create_mocks();
      mocks.totp.returns('wrong token');
      authentication.authenticate(mocks.req, mocks.res, mocks.args)
      .fail(function(err) {
        clock.restore();
        done();
      })
    });
  });


  describe('test verify authentication', function() {
    it('should be already authenticated', function(done) {
      var mocks = create_mocks();
      var data = { user: 'username' };
      mocks.jwt.verify = sinon.promise().resolves(data);
      authentication.verify_authentication(mocks.req, mocks.res, mocks.args)
      .then(function(actual_data) {
        assert.equal(actual_data, data);
        done();
      });
    });

    it('should not be already authenticated', function(done) {
      var mocks = create_mocks();
      var data = { user: 'username' };
      mocks.jwt.verify = sinon.promise().rejects('Error with JWT token');
      return authentication.verify_authentication(mocks.req, mocks.res, mocks.args)
      .fail(function() {
        done();
      });
    });
  });
});

