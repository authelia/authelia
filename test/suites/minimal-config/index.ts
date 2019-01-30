import ChildProcess from 'child_process';
import Bluebird from "bluebird";

import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import BadPassword from "./scenarii/BadPassword";
import RegisterTotp from './scenarii/RegisterTotp';
import ResetPassword from './scenarii/ResetPassword';
import TOTPValidation from './scenarii/TOTPValidation';

const execAsync = Bluebird.promisify(ChildProcess.exec);

AutheliaSuite('Minimal configuration', function() {
  this.timeout(10000);
  beforeEach(function() {
    return execAsync("cp users_database.example.yml users_database.yml");
  });

  describe('Bad password', BadPassword);
  describe('Reset password', ResetPassword);

  describe('TOTP Registration', RegisterTotp);
  describe('TOTP Validation', TOTPValidation);
});