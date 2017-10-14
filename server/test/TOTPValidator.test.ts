
import { TOTPValidator } from "../src/lib/TOTPValidator";
import Sinon = require("sinon");
import Speakeasy = require("speakeasy");

describe("test TOTP validation", function() {
  let totpValidator: TOTPValidator;
  let totpValidateStub: Sinon.SinonStub;

  beforeEach(() => {
    totpValidateStub = Sinon.stub(Speakeasy.totp, "verify");
    totpValidator = new TOTPValidator(Speakeasy);
  });

  afterEach(function() {
    totpValidateStub.restore();
  });

  it("should validate the TOTP token", function() {
    const totp_secret = "NBD2ZV64R9UV1O7K";
    const token = "token";
    totpValidateStub.withArgs({
      secret: totp_secret,
      token: token,
      encoding: "base32",
      window: 1
    }).returns(true);
    return totpValidator.validate(token, totp_secret);
  });

  it("should not validate a wrong TOTP token", function(done) {
    const totp_secret = "NBD2ZV64R9UV1O7K";
    const token = "wrong token";
    totpValidateStub.returns(false);
    totpValidator.validate(token, totp_secret)
    .catch(function() {
      done();
    });
  });
});

