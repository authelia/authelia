import * as BluebirdPromise from "bluebird";
import { IUserDataStore } from "./IUserDataStore";
import { ICollection } from "./ICollection";
import { ICollectionFactory } from "./ICollectionFactory";
import { TOTPSecretDocument } from "./TOTPSecretDocument";
import { U2FRegistrationDocument } from "./U2FRegistrationDocument";
import { U2FRegistration } from "../../../types/U2FRegistration";
import { TOTPSecret } from "../../../types/TOTPSecret";
import { AuthenticationTraceDocument } from "./AuthenticationTraceDocument";
import { IdentityValidationDocument } from "./IdentityValidationDocument";
import Method2FA from "../Method2FA";

// Constants

const IDENTITY_VALIDATION_TOKENS_COLLECTION_NAME = "identity_validation_tokens";
const AUTHENTICATION_TRACES_COLLECTION_NAME = "authentication_traces";

const U2F_REGISTRATIONS_COLLECTION_NAME = "u2f_registrations";
const TOTP_SECRETS_COLLECTION_NAME = "totp_secrets";
const PREFERED_2FA_METHOD_COLLECTION_NAME = "prefered_2fa_method";


export interface U2FRegistrationKey {
  userId: string;
  appId: string;
}

// Source

export class UserDataStore implements IUserDataStore {
  private u2fSecretCollection: ICollection;
  private identityCheckTokensCollection: ICollection;
  private authenticationTracesCollection: ICollection;
  private totpSecretCollection: ICollection;
  private prefered2faMethodCollection: ICollection;

  private collectionFactory: ICollectionFactory;

  constructor(collectionFactory: ICollectionFactory) {
    this.collectionFactory = collectionFactory;

    this.u2fSecretCollection = this.collectionFactory.build(U2F_REGISTRATIONS_COLLECTION_NAME);
    this.identityCheckTokensCollection = this.collectionFactory.build(IDENTITY_VALIDATION_TOKENS_COLLECTION_NAME);
    this.authenticationTracesCollection = this.collectionFactory.build(AUTHENTICATION_TRACES_COLLECTION_NAME);
    this.totpSecretCollection = this.collectionFactory.build(TOTP_SECRETS_COLLECTION_NAME);
    this.prefered2faMethodCollection = this.collectionFactory.build(PREFERED_2FA_METHOD_COLLECTION_NAME);
  }

  saveU2FRegistration(userId: string, appId: string, registration: U2FRegistration): BluebirdPromise<void> {
    const newDocument: U2FRegistrationDocument = {userId, appId, registration};
    const filter: U2FRegistrationKey = {userId, appId};

    return this.u2fSecretCollection.update(filter, newDocument, { upsert: true });
  }

  retrieveU2FRegistration(userId: string, appId: string): BluebirdPromise<U2FRegistrationDocument> {
    const filter: U2FRegistrationKey = { userId, appId };
    return this.u2fSecretCollection.findOne(filter);
  }

  saveAuthenticationTrace(userId: string, isAuthenticationSuccessful: boolean): BluebirdPromise<void> {
    const newDocument: AuthenticationTraceDocument = {
      userId, date: new Date(),
      isAuthenticationSuccessful: isAuthenticationSuccessful,
    };

    return this.authenticationTracesCollection.insert(newDocument);
  }

  retrieveLatestAuthenticationTraces(userId: string, count: number): BluebirdPromise<AuthenticationTraceDocument[]> {
    return this.authenticationTracesCollection.find({userId}, { date: -1 }, count);
  }

  produceIdentityValidationToken(userId: string, token: string, challenge: string, maxAge: number): BluebirdPromise<any> {
    const newDocument: IdentityValidationDocument = {
      userId, token, challenge,
      maxDate: new Date(new Date().getTime() + maxAge)
    };

    return this.identityCheckTokensCollection.insert(newDocument);
  }

  consumeIdentityValidationToken(token: string, challenge: string): BluebirdPromise<IdentityValidationDocument> {
    const that = this;
    const filter = {token, challenge};

    let identityValidationDocument: IdentityValidationDocument;

    return this.identityCheckTokensCollection.findOne(filter)
      .then(function (doc: IdentityValidationDocument) {
        if (!doc) {
          return BluebirdPromise.reject(new Error("Registration token does not exist"));
        }

        identityValidationDocument = doc;
        const current_date = new Date();
        if (current_date > doc.maxDate)
          return BluebirdPromise.reject(new Error("Registration token is not valid anymore"));

        return that.identityCheckTokensCollection.remove(filter);
      })
      .then(() => {
        return BluebirdPromise.resolve(identityValidationDocument);
      });
  }

  saveTOTPSecret(userId: string, secret: TOTPSecret): BluebirdPromise<void> {
    const doc = {userId, secret};
    return this.totpSecretCollection.update({userId}, doc, { upsert: true });
  }

  retrieveTOTPSecret(userId: string): BluebirdPromise<TOTPSecretDocument> {
    return this.totpSecretCollection.findOne({userId});
  }

  savePrefered2FAMethod(userId: string, method: Method2FA): BluebirdPromise<void> {
    const newDoc = {userId, method};
    return this.prefered2faMethodCollection.update({userId}, newDoc, {upsert: true});
  }

  retrievePrefered2FAMethod(userId: string): BluebirdPromise<Method2FA | undefined> {
    return this.prefered2faMethodCollection.findOne({userId})
      .then((doc) => {
        return (doc && doc.method) ? doc.method : undefined;
      });
  }
}
