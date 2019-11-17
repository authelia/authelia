import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import BadPassword from "./scenarii/BadPassword";
import RegisterTotp from './scenarii/RegisterTotp';
import ResetPassword from './scenarii/ResetPassword';
import TOTPValidation from './scenarii/TOTPValidation';
import BackendProtection from './scenarii/BackendProtection';
import VerifyEndpoint from './scenarii/VerifyEndpoint';
import AlreadyLoggedIn from './scenarii/AlreadyLoggedIn';
import { exec } from '../../helpers/utils/exec';
import BypassPolicy from "./scenarii/BypassPolicy";
import NoDuoPushOption from "./scenarii/NoDuoPushOption";

AutheliaSuite("/tmp/authelia/suites/Standalone/", function() {
  this.timeout(10000);
  
  beforeEach(async function() {
    await exec(`cp ${__dirname}/../../../internal/suites/Standalone/users.yml /tmp/authelia/suites/Standalone/users.yml`);
  });

  describe('Bypass policy', BypassPolicy)
  describe('Backend protection', BackendProtection);
  describe('Verify API endpoint', VerifyEndpoint);
  describe('Bad password', BadPassword);
  describe('Reset password', ResetPassword);
  describe('TOTP Registration', RegisterTotp);
  describe('TOTP Validation', TOTPValidation);
  describe('Already logged in', AlreadyLoggedIn);
  describe('No Duo Push method available', NoDuoPushOption);
});