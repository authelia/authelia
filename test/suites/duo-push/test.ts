import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import { exec } from '../../helpers/utils/exec';
import DuoPushNotification from "./scenarii/DuoPushNotification";
import Prefered2faMethod from "./scenarii/Prefered2faMethod";

// required to query duo-api over https
process.env["NODE_TLS_REJECT_UNAUTHORIZED"] = 0 as any;

AutheliaSuite(__dirname, function() {
  this.timeout(10000);
  
  beforeEach(async function() {
    await exec(`cp ${__dirname}/users_database.yml ${__dirname}/users_database.test.yml`);
  });

  describe("Duo Push Notication", DuoPushNotification);
  describe("Prefered 2FA methods", Prefered2faMethod);
});