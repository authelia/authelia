import WithAutheliaRunning from "./WithAutheliaRunning";
import WithDriver from "./WithDriver";

let running = false;

interface AutheliaSuiteType {
  (description: string, configPath: string, cb: (this: Mocha.ISuiteCallbackContext) => void): Mocha.ISuite;
  only: (description: string, configPath: string, cb: (this: Mocha.ISuiteCallbackContext) => void) => Mocha.ISuite;
}

function AutheliaSuiteBase(description: string, configPath: string,
  cb: (this: Mocha.ISuiteCallbackContext) => void,
  context: (description: string, ctx: (this: Mocha.ISuiteCallbackContext) => void) => Mocha.ISuite) {
  if (!running  && process.env['WITH_SERVER'] == 'y') {
    WithAutheliaRunning(configPath);
    running = true;
  }  

  return context('Suite: ' + description, function(this: Mocha.ISuiteCallbackContext) {
    WithDriver.call(this);
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