
const totp = require("../../src/lib/totp");
const sinon = require("sinon");
import Promise = require("bluebird");

describe("test TOTP validation", function() {
  it("should validate the TOTP token", function() {
    const totp_secret = "NBD2ZV64R9UV1O7K";
    const token = "token";
    const totp_mock = sinon.mock();
    totp_mock.returns("token");
    const speakeasy_mock = {
      totp: totp_mock
    };
    return totp.validate(speakeasy_mock, token, totp_secret);
  });

  it("should not validate a wrong TOTP token", function() {
    const totp_secret = "NBD2ZV64R9UV1O7K";
    const token = "wrong token";
    const totp_mock = sinon.mock();
    totp_mock.returns("token");
    const speakeasy_mock = {
      totp: totp_mock
    };
    return totp.validate(speakeasy_mock, token, totp_secret)
    .catch(function() {
      return Promise.resolve();
    });
  });
});

