import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import MongoConnectionRecovery from "./scenarii/MongoConnectionRecovery";
import EnforceInternalRedirectionsOnly from "./scenarii/EnforceInternalRedirectionsOnly";
import AccessControl from "./scenarii/AccessControl";

AutheliaSuite('Complete configuration', __dirname + '/config.yml', function() {
  this.timeout(10000);

  describe('Access control', AccessControl);

  describe('Mongo broken connection recovery', MongoConnectionRecovery);
  describe('Enforce internal redirections only', EnforceInternalRedirectionsOnly);
});