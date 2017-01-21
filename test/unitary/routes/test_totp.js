
var totp = require('../../../src/lib/routes/totp');
var Promise = require('bluebird');
var sinon = require('sinon');
var assert = require('assert');

describe('test totp route', function() {
  var req, res;
  var totp_engine;

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
    totp_engine = {
      totp: sinon.stub()
    }
    app_get.withArgs('totp engine').returns(totp_engine);
    app_get.withArgs('config').returns(config);
  });


  it('should send status code 204 when totp is valid', function() {
    return new Promise(function(resolve, reject) {
      totp_engine.totp.returns('abc');
      res.send = sinon.spy(function() {
        // Second factor passed
        assert.equal(true, req.session.auth_session.second_factor)
        assert.equal(204, res.status.getCall(0).args[0]);
        resolve();
      });
      totp(req, res); 
    })
  });

  it('should send status code 401 when totp is not valid', function() {
    return new Promise(function(resolve, reject) {
      totp_engine.totp.returns('bad_token');
      res.send = sinon.spy(function() {
        assert.equal(false, req.session.auth_session.second_factor)
        assert.equal(401, res.status.getCall(0).args[0]);
        resolve();
      });
      totp(req, res); 
    })
  });

  it('should send status code 401 when session has not been initiated', function() {
    return new Promise(function(resolve, reject) {
      totp_engine.totp.returns('abc');
      res.send = sinon.spy(function() {
        assert.equal(401, res.status.getCall(0).args[0]);
        resolve();
      });
      req.session = {};
      totp(req, res); 
    })
  });
});

