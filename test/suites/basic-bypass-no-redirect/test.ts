import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import { exec } from '../../helpers/utils/exec';
import BypassPolicy from "./scenarii/BypassPolicy";
import NoDefaultRedirectionUrl from "./scenarii/NoDefaultRedirectionUrl";
import CustomHeadersForwarded from "./scenarii/CustomHeadersForwarded";

process.env["NODE_TLS_REJECT_UNAUTHORIZED"] = 0 as any;

process.env["NODE_TLS_REJECT_UNAUTHORIZED"] = 0 as any;

AutheliaSuite(__dirname, function() {
  this.timeout(10000);
  
  beforeEach(async function() {
    await exec(`cp ${__dirname}/users_database.yml ${__dirname}/users_database.test.yml`);
  });

  describe('Bypass policy', BypassPolicy);
  describe("No default redirection", NoDefaultRedirectionUrl);
  describe("Custom headers forwarded on bypass", CustomHeadersForwarded);
});