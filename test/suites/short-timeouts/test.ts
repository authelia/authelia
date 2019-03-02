import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import Inactivity from './scenarii/Inactivity';
import { exec } from '../../helpers/utils/exec';

AutheliaSuite('Short timeouts', __dirname, function() {
  this.timeout(10000);
  beforeEach(async function() {
    await exec('cp users_database.example.yml users_database.yml');
  });

  describe('Inactivity period', Inactivity);
});