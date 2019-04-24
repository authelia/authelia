import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import SingleFactorAuthentication from "../../helpers/scenarii/SingleFactorAuthentication";
import TwoFactorAuthentication from "../../helpers/scenarii/TwoFactorAuthentication";
import AuthenticationBlacklisting from "../../helpers/scenarii/AuthenticationBlacklisting";

AutheliaSuite(__dirname, function() {
  this.timeout(20000);

  describe('Single-factor authentication', SingleFactorAuthentication())
  describe('Two-factor authentication', TwoFactorAuthentication());
  describe('Authentication blacklisting', AuthenticationBlacklisting(12000));
});