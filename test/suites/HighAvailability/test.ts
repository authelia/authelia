import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import MariaConnectionRecovery from "./scenarii/MariaConnectionRecovery";
import EnforceInternalRedirectionsOnly from "./scenarii/EnforceInternalRedirectionsOnly";
import AccessControl from "./scenarii/AccessControl";
import CustomHeadersForwarded from "./scenarii/CustomHeadersForwarded";
import BasicAuthentication from "./scenarii/BasicAuthentication";
import AutheliaRestart from "./scenarii/AutheliaRestart";
import AuthenticationRegulation from "./scenarii/AuthenticationRegulation";

AutheliaSuite(__dirname, function () {
  this.timeout(10000);

  describe('Custom headers forwarded to backend', CustomHeadersForwarded);
  describe('Access control', AccessControl);
  describe('Mariadb broken connection recovery', MariaConnectionRecovery);
  describe('Enforce internal redirections only', EnforceInternalRedirectionsOnly);
  describe('Basic authentication', BasicAuthentication);
  describe('Authelia restart', AutheliaRestart);
  describe('Authentication regulation', AuthenticationRegulation);
});