import WithAutheliaRunning from "./WithAutheliaRunning";
import WithDriver from "./WithDriver";

let running = false;

interface AutheliaSuiteType {
  (description: string, cb: (this: Mocha.ISuiteCallbackContext) => void): Mocha.ISuite;
  only: (description: string, cb: (this: Mocha.ISuiteCallbackContext) => void) => Mocha.ISuite;
}

function AutheliaSuiteBase(description: string,
  context: (description: string, ctx: (this: Mocha.ISuiteCallbackContext) => void) => Mocha.ISuite,
  cb: (this: Mocha.ISuiteCallbackContext) => void) {
  if (!running  && process.env['WITH_SERVER'] == 'y') {
    WithAutheliaRunning();
    running = true;
  }  

  return context('Suite: ' + description, function(this: Mocha.ISuiteCallbackContext) {
    WithDriver.call(this);
    cb.call(this);
  });
}

const AutheliaSuite = <AutheliaSuiteType>function(description: string, cb: (this: Mocha.ISuiteCallbackContext) => void) {
  return AutheliaSuiteBase(description, describe, cb);
}


AutheliaSuite.only = function(description: string, cb: (this: Mocha.ISuiteCallbackContext) => void) {
  return AutheliaSuiteBase(description, describe.only, cb);
}

export default AutheliaSuite as AutheliaSuiteType;