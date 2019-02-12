import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import MongoConnectionRecovery from "./scenarii/MongoConnectionRecovery";
import EnforceInternalRedirectionsOnly from "./scenarii/EnforceInternalRedirectionsOnly";
import AccessControl from "./scenarii/AccessControl";
import CustomHeadersForwarded from "./scenarii/CustomHeadersForwarded";

AutheliaSuite('Complete configuration', __dirname + '/config.yml', function() {
  this.timeout(10000);

  describe('Custom headers forwarded to backend', CustomHeadersForwarded);
  describe('Access control', AccessControl);

  describe('Mongo broken connection recovery', MongoConnectionRecovery);
  describe('Enforce internal redirections only', EnforceInternalRedirectionsOnly);
});