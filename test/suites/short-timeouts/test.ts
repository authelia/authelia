import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import Inactivity from './scenarii/Inactivity';

AutheliaSuite(__dirname, function() {
  this.timeout(10000);
  describe('Inactivity period', Inactivity);
});