import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import MongoConnectionRecovery from "./scenarii/MongoConnectionRecovery";
import EnforceInternalRedirectionsOnly from "./scenarii/EnforceInternalRedirectionsOnly";
import AccessControl from "./scenarii/AccessControl";
import CustomHeadersForwarded from "./scenarii/CustomHeadersForwarded";
import BasicAuthentication from "./scenarii/BasicAuthentication";
import AutheliaRestart from "./scenarii/AutheliaRestart";
import AuthenticationRegulation from "./scenarii/AuthenticationRegulation";
import SingleFactorAuthentication from "../../helpers/scenarii/SingleFactorAuthentication";

AutheliaSuite(__dirname, function() {
  this.timeout(10000);

  describe('Custom headers forwarded to backend', CustomHeadersForwarded);
  describe('Access control', AccessControl);
  describe('Mongo broken connection recovery', MongoConnectionRecovery);
  describe('Enforce internal redirections only', EnforceInternalRedirectionsOnly);
  describe('Single-factor authentication', SingleFactorAuthentication());
  describe('Basic authentication', BasicAuthentication);
  describe('Authelia restart', AutheliaRestart);
  describe('Authentication regulation', AuthenticationRegulation);
});