import AutheliaSuite from '../../helpers/context/AutheliaSuite';
import TwoFactorAuthentication from '../../helpers/scenarii/TwoFactorAuthentication';
import SingleFactorAuthentication from '../../helpers/scenarii/SingleFactorAuthentication';

AutheliaSuite(__dirname, function() {
  this.timeout(30000);

  describe('Single-factor authentication', SingleFactorAuthentication(30000));
  describe('Two-factor authentication', TwoFactorAuthentication(30000));
});