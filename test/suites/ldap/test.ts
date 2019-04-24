import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import SingleFactorAuthentication from "../../helpers/scenarii/SingleFactorAuthentication";
import TwoFactorAuthentication from "../../helpers/scenarii/TwoFactorAuthentication";

AutheliaSuite(__dirname, function() {
  this.timeout(10000);

  describe('Single-factor authentication', SingleFactorAuthentication())
  describe('Two-factor authentication', TwoFactorAuthentication());
});