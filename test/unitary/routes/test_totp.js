
var totp = require('../../../src/lib/routes/totp');
var Promise = require('bluebird');
var sinon = require('sinon');
var assert = require('assert');
var winston = require('winston');

describe('test totp route', function() {
  var req, res;
  var totpValidator;
  var user_data_store;

  beforeEach(function() {
    var app_get = sinon.stub();
    req = {
      app: {
        get: app_get
      },
      body: {
        token: 'abc'
      },
      session: {
        auth_session: {
          userid: 'user',
          first_factor: false,
          second_factor: false
        }
      }
    };
    res = {
      send: sinon.spy(),
      status: sinon.spy()
    };

    var config = { totp_secret: 'secret' };
    totpValidator = {
      validate: sinon.stub()
    }

    user_data_store = {};
    user_data_store.get_totp_secret = sinon.stub();

    var doc = {};
    doc.userid = 'user';
    doc.secret = {};
    doc.secret.base32 = 'ABCDEF';
    user_data_store.get_totp_secret.returns(Promise.resolve(doc));

    app_get.withArgs('logger').returns(winston);
    app_get.withArgs('totp validator').returns(totpValidator);
    app_get.withArgs('config').returns(config);
    app_get.withArgs('user data store').returns(user_data_store);
  });


  it('should send status code 204 when totp is valid', function(done) {
    totpValidator.validate.returns(Promise.resolve("ok"));
    res.send = sinon.spy(function() {
      // Second factor passed
      assert.equal(true, req.session.auth_session.second_factor)
      assert.equal(204, res.status.getCall(0).args[0]);
      done();
    });
    totp(req, res); 
  });

  it('should send status code 401 when totp is not valid', function(done) {
    totpValidator.validate.returns(Promise.reject('bad_token'));
    res.send = sinon.spy(function() {
      assert.equal(false, req.session.auth_session.second_factor)
      assert.equal(401, res.status.getCall(0).args[0]);
      done();
    });
    totp(req, res); 
  });

  it('should send status code 401 when session has not been initiated', function(done) {
    totpValidator.validate.returns(Promise.resolve('abc'));
    res.send = sinon.spy(function() {
      assert.equal(403, res.status.getCall(0).args[0]);
      done();
    });
    req.session = {};
    totp(req, res); 
  });
});

