import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");

import { TOTPSecretDocument } from "../../../src/lib/storage/TOTPSecretDocument";
import { U2FRegistrationDocument } from "../../../src/lib/storage/U2FRegistrationDocument";
import { U2FRegistration } from "../../../types/U2FRegistration";
import { TOTPSecret } from "../../../types/TOTPSecret";
import { AuthenticationTraceDocument } from "../../../src/lib/storage/AuthenticationTraceDocument";
import { IdentityValidationDocument } from "../../../src/lib/storage/IdentityValidationDocument";

import { IUserDataStore } from "../../../src/lib/storage/IUserDataStore";

export class UserDataStoreStub implements IUserDataStore {
    saveU2FRegistrationStub: Sinon.SinonStub;
    retrieveU2FRegistrationStub: Sinon.SinonStub;
    saveAuthenticationTraceStub: Sinon.SinonStub;
    retrieveLatestAuthenticationTracesStub: Sinon.SinonStub;
    produceIdentityValidationTokenStub: Sinon.SinonStub;
    consumeIdentityValidationTokenStub: Sinon.SinonStub;
    saveTOTPSecretStub: Sinon.SinonStub;
    retrieveTOTPSecretStub: Sinon.SinonStub;

    constructor() {
        this.saveU2FRegistrationStub = Sinon.stub();
        this.retrieveU2FRegistrationStub = Sinon.stub();
        this.saveAuthenticationTraceStub = Sinon.stub();
        this.retrieveLatestAuthenticationTracesStub = Sinon.stub();
        this.produceIdentityValidationTokenStub = Sinon.stub();
        this.consumeIdentityValidationTokenStub = Sinon.stub();
        this.saveTOTPSecretStub = Sinon.stub();
        this.retrieveTOTPSecretStub = Sinon.stub();
    }

    saveU2FRegistration(userId: string, appId: string, registration: U2FRegistration): BluebirdPromise<void> {
        return this.saveU2FRegistrationStub(userId, appId, registration);
    }

    retrieveU2FRegistration(userId: string, appId: string): BluebirdPromise<U2FRegistrationDocument> {
        return this.retrieveU2FRegistrationStub(userId, appId);
    }

    saveAuthenticationTrace(userId: string, isAuthenticationSuccessful: boolean): BluebirdPromise<void> {
        return this.saveAuthenticationTraceStub(userId, isAuthenticationSuccessful);
    }

    retrieveLatestAuthenticationTraces(userId: string, count: number): BluebirdPromise<AuthenticationTraceDocument[]> {
        return this.retrieveLatestAuthenticationTracesStub(userId, count);
    }

    produceIdentityValidationToken(userId: string, token: string, challenge: string, maxAge: number): BluebirdPromise<any> {
        return this.produceIdentityValidationTokenStub(userId, token, challenge, maxAge);
    }

    consumeIdentityValidationToken(token: string, challenge: string): BluebirdPromise<IdentityValidationDocument> {
        return this.consumeIdentityValidationTokenStub(token, challenge);
    }

    saveTOTPSecret(userId: string, secret: TOTPSecret): BluebirdPromise<void> {
        return this.saveTOTPSecretStub(userId, secret);
    }

    retrieveTOTPSecret(userId: string): BluebirdPromise<TOTPSecretDocument> {
        return this.retrieveTOTPSecretStub(userId);
    }
}