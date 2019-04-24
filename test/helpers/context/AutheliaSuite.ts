interface AutheliaSuiteType {
  (suitePath: string, cb: (this: Mocha.ISuiteCallbackContext) => void): Mocha.ISuite;
  only: (suitePath: string, cb: (this: Mocha.ISuiteCallbackContext) => void) => Mocha.ISuite;
}

function AutheliaSuiteBase(suitePath: string,
  cb: (this: Mocha.ISuiteCallbackContext) => void,
  context: (description: string, ctx: (this: Mocha.ISuiteCallbackContext) => void) => Mocha.ISuite) {
  const suite = suitePath.split('/').slice(-1)[0];
  return context('Suite: ' + suite, function(this: Mocha.ISuiteCallbackContext) {
    cb.call(this);
  });
}

const AutheliaSuite = <AutheliaSuiteType>function(suitePath: string, 
  cb: (this: Mocha.ISuiteCallbackContext) => void) {
  return AutheliaSuiteBase(suitePath, cb, describe);
}


AutheliaSuite.only = function(suitePath: string, 
  cb: (this: Mocha.ISuiteCallbackContext) => void) {
  return AutheliaSuiteBase(suitePath, cb, describe.only);
}

export default AutheliaSuite as AutheliaSuiteType;