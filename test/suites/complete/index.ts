import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import MongoConnectionRecovery from "./scenarii/MongoConnectionRecovery";
import EnforceInternalRedirectionsOnly from "./scenarii/EnforceInternalRedirectionsOnly";

AutheliaSuite('Complete configuration', __dirname + 'config.yml', function() {
  this.timeout(10000);

  describe('Mongo broken connection recovery', MongoConnectionRecovery);
  describe('Enforce internal redirections only', EnforceInternalRedirectionsOnly);
});