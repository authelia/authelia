import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import Inactivity from './scenarii/Inactivity';

AutheliaSuite('Short timeouts', __dirname, function() {
  this.timeout(10000);
  describe('Inactivity period', Inactivity);
});