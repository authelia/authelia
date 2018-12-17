import U2f = require("u2f");

export interface IU2fHandler {
  request(appId: string, keyHandle?: string): U2f.Request;
  checkRegistration(registrationRequest: U2f.Request, registrationResponse: U2f.RegistrationData)
    : U2f.RegistrationResult | U2f.Error;
  checkSignature(signatureRequest: U2f.Request, signatureResponse: U2f.SignatureData, publicKey: string)
    : U2f.SignatureResult | U2f.Error;
}