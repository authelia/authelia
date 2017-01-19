
var assert = require('assert');
var authentication = require('../../src/lib/authentication');
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
    },
    app: {
      get: sinon.stub()
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

  req_mock.app.get.withArgs('ldap client').returns(args.ldap_interface);
  req_mock.app.get.withArgs('jwt engine').returns(args.jwt);
  req_mock.app.get.withArgs('totp engine').returns(args.totp_interface);
  req_mock.app.get.withArgs('config').returns({
    totp_secret: 'totp_secret',
    ldap_users_dn: 'ou=users,dc=example,dc=com'
  });

  return {
    req: req_mock, 
    res: res_mock, 
    args: args,
    totp: totp_mock,
    jwt: jwt
  }
}

describe('test authentication token verification', function() {
  it('should be already authenticated', function(done) {
    var mocks = create_mocks();
    var data = { user: 'username' };
    mocks.req.app.get.withArgs('jwt engine').returns({
      verify: sinon.promise().resolves(data)
    });

    authentication.verify(mocks.req, mocks.res)
    .then(function(actual_data) {
      assert.equal(actual_data, data);
      done();
    });
  });

  it('should not be already authenticated', function(done) {
    var mocks = create_mocks();
    var data = { user: 'username' };
    mocks.req.app.get.withArgs('jwt engine').returns({
      verify: sinon.promise().rejects('Error with JWT token')
    });
    return authentication.verify(mocks.req, mocks.res, mocks.args)
    .fail(function() {
      done();
    });
  });
});

