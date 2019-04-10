import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import { exec } from '../../helpers/utils/exec';
import TwoFactorAuthentication from "../../helpers/scenarii/TwoFactorAuthentication";
import SingleFactorAuthentication from "../../helpers/scenarii/SingleFactorAuthentication";
import * as fs from "fs";

AutheliaSuite(__dirname, function() {
  this.timeout(10000);
  
  beforeEach(async function() {
    await exec('./example/compose/traefik/render.js ' + (fs.existsSync('.suite') ? '': '--production'));
    await exec(`cp ${__dirname}/users_database.yml ${__dirname}/users_database.test.yml`);
  });

  describe('Single-factor authentication', SingleFactorAuthentication());
  describe('Second factor authentication', TwoFactorAuthentication());
});