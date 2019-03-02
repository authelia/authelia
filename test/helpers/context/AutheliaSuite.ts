import WithEnvironment from "./WithEnvironment";
import fs from 'fs';

interface AutheliaSuiteType {
  (description: string, configPath: string, cb: (this: Mocha.ISuiteCallbackContext) => void): Mocha.ISuite;
  only: (description: string, configPath: string, cb: (this: Mocha.ISuiteCallbackContext) => void) => Mocha.ISuite;
}

function AutheliaSuiteBase(description: string, suite: string,
  cb: (this: Mocha.ISuiteCallbackContext) => void,
  context: (description: string, ctx: (this: Mocha.ISuiteCallbackContext) => void) => Mocha.ISuite) {
  return context('Suite: ' + description, function(this: Mocha.ISuiteCallbackContext) {
    if (!fs.existsSync('.suite')) {
      WithEnvironment(suite);
    }

    cb.call(this);
  });
}

const AutheliaSuite = <AutheliaSuiteType>function(
  description: string, configPath: string, 
  cb: (this: Mocha.ISuiteCallbackContext) => void) {
  return AutheliaSuiteBase(description, configPath, cb, describe);
}


AutheliaSuite.only = function(description: string, configPath: string, 
  cb: (this: Mocha.ISuiteCallbackContext) => void) {
  return AutheliaSuiteBase(description, configPath, cb, describe.only);
}

export default AutheliaSuite as AutheliaSuiteType;