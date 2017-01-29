
var totp = require('../../src/lib/totp');
var sinon = require('sinon');
var Promise = require('bluebird');

describe('test TOTP validation', function() {
  it('should validate the TOTP token', function() {
    var totp_secret = 'NBD2ZV64R9UV1O7K';
    var token = 'token';
    var totp_mock = sinon.mock();
    totp_mock.returns('token');
    var speakeasy_mock = {
      totp: totp_mock
    }
    return totp.validate(speakeasy_mock, token, totp_secret);
  });

  it('should not validate a wrong TOTP token', function() {
    var totp_secret = 'NBD2ZV64R9UV1O7K';
    var token = 'wrong token';
    var totp_mock = sinon.mock();
    totp_mock.returns('token');
    var speakeasy_mock = {
      totp: totp_mock
    }
    return totp.validate(speakeasy_mock, token, totp_secret)
    .catch(function() {
      return Promise.resolve();
    });
  });
});

