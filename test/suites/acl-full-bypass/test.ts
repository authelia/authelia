import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import { exec } from '../../helpers/utils/exec';
import BypassPolicy from "./scenarii/BypassPolicy";

AutheliaSuite(__dirname, function() {
  this.timeout(10000);
  
  beforeEach(async function() {
    await exec(`cp ${__dirname}/users_database.yml ${__dirname}/users_database.test.yml`);
  });

  describe('Bypass policy', BypassPolicy);
});