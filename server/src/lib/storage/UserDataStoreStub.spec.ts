import * as Sinon from "sinon";
import * as BluebirdPromise from "bluebird";

import { TOTPSecretDocument } from "./TOTPSecretDocument";
import { U2FRegistrationDocument } from "./U2FRegistrationDocument";
import { U2FRegistration } from "../../../types/U2FRegistration";
import { TOTPSecret } from "../../../types/TOTPSecret";
import { AuthenticationTraceDocument } from "./AuthenticationTraceDocument";
import { IdentityValidationDocument } from "./IdentityValidationDocument";
import { IUserDataStore } from "./IUserDataStore";
import Method2FA from "../Method2FA";

export class UserDataStoreStub implements IUserDataStore {
    saveU2FRegistrationStub: Sinon.SinonStub;
    retrieveU2FRegistrationStub: Sinon.SinonStub;
    saveAuthenticationTraceStub: Sinon.SinonStub;
    retrieveLatestAuthenticationTracesStub: Sinon.SinonStub;
    produceIdentityValidationTokenStub: Sinon.SinonStub;
    consumeIdentityValidationTokenStub: Sinon.SinonStub;
    saveTOTPSecretStub: Sinon.SinonStub;
    retrieveTOTPSecretStub: Sinon.SinonStub;
    savePrefered2FAMethodStub: Sinon.SinonStub;
    retrievePrefered2FAMethodStub: Sinon.SinonStub;

    constructor() {
        this.saveU2FRegistrationStub = Sinon.stub();
        this.retrieveU2FRegistrationStub = Sinon.stub();
        this.saveAuthenticationTraceStub = Sinon.stub();
        this.retrieveLatestAuthenticationTracesStub = Sinon.stub();
        this.produceIdentityValidationTokenStub = Sinon.stub();
        this.consumeIdentityValidationTokenStub = Sinon.stub();
        this.saveTOTPSecretStub = Sinon.stub();
        this.retrieveTOTPSecretStub = Sinon.stub();
        this.savePrefered2FAMethodStub = Sinon.stub();
        this.retrievePrefered2FAMethodStub = Sinon.stub();
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

    savePrefered2FAMethod(userId: string, method: Method2FA): BluebirdPromise<void> {
        return this.savePrefered2FAMethodStub(userId, method);
    }

    retrievePrefered2FAMethod(userId: string): BluebirdPromise<Method2FA> {
        return this.retrievePrefered2FAMethodStub(userId);
    }
}