import ChildProcess from 'child_process';
import Bluebird from "bluebird";

import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import BadPassword from "./scenarii/BadPassword";
import RegisterTotp from './scenarii/RegisterTotp';
import ResetPassword from './scenarii/ResetPassword';
import TOTPValidation from './scenarii/TOTPValidation';
import Inactivity from './scenarii/Inactivity';
import BackendProtection from './scenarii/BackendProtection';
import VerifyEndpoint from './scenarii/VerifyEndpoint';

const execAsync = Bluebird.promisify(ChildProcess.exec);

AutheliaSuite('Minimal configuration', __dirname + '/config.yml', function() {
  this.timeout(10000);
  beforeEach(function() {
    return execAsync("cp users_database.example.yml users_database.yml");
  });

  describe('Backend protection', BackendProtection);
  describe('Verify API endpoint', VerifyEndpoint);

  describe('Bad password', BadPassword);
  describe('Reset password', ResetPassword);

  describe('TOTP Registration', RegisterTotp);
  describe('TOTP Validation', TOTPValidation);

  describe('Inactivity period', Inactivity);
});