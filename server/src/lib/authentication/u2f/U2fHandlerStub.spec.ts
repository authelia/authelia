import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import U2f = require("u2f");
import { IU2fHandler } from "./IU2fHandler";


export class U2fHandlerStub implements IU2fHandler {
  requestStub: Sinon.SinonStub;
  checkRegistrationStub: Sinon.SinonStub;
  checkSignatureStub: Sinon.SinonStub;

  constructor() {
    this.requestStub = Sinon.stub();
    this.checkRegistrationStub = Sinon.stub();
    this.checkSignatureStub = Sinon.stub();
  }

  request(appId: string, keyHandle?: string): U2f.Request {
    return this.requestStub(appId, keyHandle);
  }

  checkRegistration(registrationRequest: U2f.Request, registrationResponse: U2f.RegistrationData)
    : U2f.RegistrationResult | U2f.Error {
    return this.checkRegistrationStub(registrationRequest, registrationResponse);
  }

  checkSignature(signatureRequest: U2f.Request, signatureResponse: U2f.SignatureData, publicKey: string)
    : U2f.SignatureResult | U2f.Error {
    return this.checkSignatureStub(signatureRequest, signatureResponse, publicKey);
  }
}