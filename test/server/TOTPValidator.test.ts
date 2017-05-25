
import { TOTPValidator } from "../../src/server/lib/TOTPValidator";
import sinon = require("sinon");
import Promise = require("bluebird");
import SpeakeasyMock = require("./mocks/speakeasy");

describe("test TOTP validation", function() {
  let totpValidator: TOTPValidator;

  beforeEach(() => {
    SpeakeasyMock.totp.returns("token");
    totpValidator = new TOTPValidator(SpeakeasyMock as any);
  });

  it("should validate the TOTP token", function() {
    const totp_secret = "NBD2ZV64R9UV1O7K";
    const token = "token";
    return totpValidator.validate(token, totp_secret);
  });

  it("should not validate a wrong TOTP token", function(done) {
    const totp_secret = "NBD2ZV64R9UV1O7K";
    const token = "wrong token";
    totpValidator.validate(token, totp_secret)
    .catch(function() {
      done();
    });
  });
});

