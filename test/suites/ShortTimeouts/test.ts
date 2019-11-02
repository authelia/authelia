import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import Inactivity from './scenarii/Inactivity';
import AuthenticationBlacklisting from "../../helpers/scenarii/AuthenticationBlacklisting";

AutheliaSuite(__dirname, function () {
  this.timeout(10000);
  describe('Inactivity period', Inactivity);
  describe('Authentication blacklisting', AuthenticationBlacklisting(10000));
});