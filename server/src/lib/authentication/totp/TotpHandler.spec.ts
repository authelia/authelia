
import { TotpHandler } from "./TotpHandler";
import Sinon = require("sinon");
import Speakeasy = require("speakeasy");
import Assert = require("assert");

describe("authentication/totp/TotpHandler", function() {
  let totpValidator: TotpHandler;
  let validateStub: Sinon.SinonStub;

  beforeEach(() => {
    validateStub = Sinon.stub(Speakeasy.totp, "verify");
    totpValidator = new TotpHandler(Speakeasy);
  });

  afterEach(function() {
    validateStub.restore();
  });

  it("should validate the TOTP token", function() {
    const totp_secret = "NBD2ZV64R9UV1O7K";
    const token = "token";
    validateStub.withArgs({
      secret: totp_secret,
      token: token,
      encoding: "base32",
      window: 1
    }).returns(true);
    Assert(totpValidator.validate(token, totp_secret));
  });

  it("should not validate a wrong TOTP token", function() {
    const totp_secret = "NBD2ZV64R9UV1O7K";
    const token = "wrong token";
    validateStub.returns(false);
    Assert(!totpValidator.validate(token, totp_secret));
  });
});

