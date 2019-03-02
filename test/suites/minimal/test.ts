import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import BadPassword from "./scenarii/BadPassword";
import RegisterTotp from './scenarii/RegisterTotp';
import ResetPassword from './scenarii/ResetPassword';
import TOTPValidation from './scenarii/TOTPValidation';
import Inactivity from './scenarii/Inactivity';
import BackendProtection from './scenarii/BackendProtection';
import VerifyEndpoint from './scenarii/VerifyEndpoint';
import RequiredTwoFactor from './scenarii/RequiredTwoFactor';
import LogoutRedirectToAlreadyLoggedIn from './scenarii/LogoutRedirectToAlreadyLoggedIn';
import SimpleAuthentication from './scenarii/SimpleAuthentication';
import { exec } from '../../helpers/utils/exec';

AutheliaSuite('Minimal configuration', __dirname, function() {
  this.timeout(10000);
  beforeEach(async function() {
    await exec('cp users_database.example.yml users_database.yml');
  });

  describe('Simple authentication', SimpleAuthentication);
  describe('Backend protection', BackendProtection);
  describe('Verify API endpoint', VerifyEndpoint);
  describe('Bad password', BadPassword);
  describe('Reset password', ResetPassword);
  describe('TOTP Registration', RegisterTotp);
  describe('TOTP Validation', TOTPValidation);
  describe('Inactivity period', Inactivity);
  describe('Required two factor', RequiredTwoFactor);
  describe('Logout endpoint redirect to already logged in page', LogoutRedirectToAlreadyLoggedIn);
});