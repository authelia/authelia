import { IU2fHandler } from "./IU2fHandler";
import U2f = require("u2f");

export class U2fHandler implements IU2fHandler {
  private u2f: typeof U2f;

  constructor(u2f: typeof U2f) {
    this.u2f = u2f;
  }

  request(appId: string, keyHandle?: string): U2f.Request {
    return this.u2f.request(appId, keyHandle);
  }

  checkRegistration(registrationRequest: U2f.Request, registrationResponse: U2f.RegistrationData)
    : U2f.RegistrationResult | U2f.Error {
    return this.u2f.checkRegistration(registrationRequest, registrationResponse);
  }

  checkSignature(signatureRequest: U2f.Request, signatureResponse: U2f.SignatureData, publicKey: string)
    : U2f.SignatureResult | U2f.Error {
    return this.u2f.checkSignature(signatureRequest, signatureResponse, publicKey);
  }
}
