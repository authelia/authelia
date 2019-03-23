
import * as Assert from "assert";
import * as Sinon from "sinon";
import * as MockDate from "mockdate";
import BluebirdPromise = require("bluebird");

import { UserDataStore } from "./UserDataStore";
import { TOTPSecret } from "../../../types/TOTPSecret";
import { U2FRegistration } from "../../../types/U2FRegistration";
import { AuthenticationTraceDocument } from "./AuthenticationTraceDocument";
import { CollectionStub } from "./CollectionStub.spec";
import { CollectionFactoryStub } from "./CollectionFactoryStub.spec";

describe("storage/UserDataStore", function () {
  let factory: CollectionFactoryStub;
  let collection: CollectionStub;
  let userId: string;
  let appId: string;
  let totpSecret: TOTPSecret;
  let u2fRegistration: U2FRegistration;

  beforeEach(function () {
    factory = new CollectionFactoryStub();
    collection = new CollectionStub();

    userId = "user";
    appId = "https://myappId";

    totpSecret = {
      ascii: "abc",
      base32: "ABCDKZLEFZGREJK",
      otpauth_url: "totp://test",
      google_auth_qr: "dummy",
      hex: "dummy",
      qr_code_ascii: "dummy",
      qr_code_base32: "dummy",
      qr_code_hex: "dummy"
    };

    u2fRegistration = {
      keyHandle: "KEY_HANDLE",
      publicKey: "publickey"
    };
  });

  it("should correctly creates collections", function () {
    new UserDataStore(factory);

    Assert.equal(5, factory.buildStub.callCount);
    Assert(factory.buildStub.calledWith("authentication_traces"));
    Assert(factory.buildStub.calledWith("identity_validation_tokens"));
    Assert(factory.buildStub.calledWith("u2f_registrations"));
    Assert(factory.buildStub.calledWith("totp_secrets"));
    Assert(factory.buildStub.calledWith("prefered_2fa_method"));
  });

  describe("TOTP secrets collection", function () {
    it("should save a totp secret", function () {
      factory.buildStub.returns(collection);
      collection.updateStub.returns(BluebirdPromise.resolve());

      const dataStore = new UserDataStore(factory);

      return dataStore.saveTOTPSecret(userId, totpSecret)
        .then(function (doc) {
          Assert(collection.updateStub.calledOnce);
          Assert(collection.updateStub.calledWith({ userId: userId }, {
            userId: userId,
            secret: totpSecret
          }, { upsert: true }));
          return BluebirdPromise.resolve();
        });
    });

    it("should retrieve a totp secret", function () {
      factory.buildStub.returns(collection);
      collection.findOneStub.withArgs().returns(BluebirdPromise.resolve());

      const dataStore = new UserDataStore(factory);

      return dataStore.retrieveTOTPSecret(userId)
        .then(function (doc) {
          Assert(collection.findOneStub.calledOnce);
          Assert(collection.findOneStub.calledWith({ userId: userId }));
          return BluebirdPromise.resolve();
        });
    });
  });

  describe("U2F secrets collection", function () {
    it("should save a U2F secret", function () {
      factory.buildStub.returns(collection);
      collection.updateStub.returns(BluebirdPromise.resolve());

      const dataStore = new UserDataStore(factory);

      return dataStore.saveU2FRegistration(userId, appId, u2fRegistration)
        .then(function (doc) {
          Assert(collection.updateStub.calledOnce);
          Assert(collection.updateStub.calledWith({
            userId: userId,
            appId: appId
          }, {
              userId: userId,
              appId: appId,
              registration: u2fRegistration
            }, { upsert: true }));
          return BluebirdPromise.resolve();
        });
    });

    it("should retrieve a U2F secret", function () {
      factory.buildStub.returns(collection);
      collection.findOneStub.withArgs().returns(BluebirdPromise.resolve());

      const dataStore = new UserDataStore(factory);

      return dataStore.retrieveU2FRegistration(userId, appId)
        .then(function (doc) {
          Assert(collection.findOneStub.calledOnce);
          Assert(collection.findOneStub.calledWith({
            userId: userId,
            appId: appId
          }));
          return BluebirdPromise.resolve();
        });
    });
  });


  describe("Regulator traces collection", function () {
    it("should save a trace", function () {
      factory.buildStub.returns(collection);
      collection.insertStub.returns(BluebirdPromise.resolve());

      const dataStore = new UserDataStore(factory);

      return dataStore.saveAuthenticationTrace(userId, true)
        .then(function (doc) {
          Assert(collection.insertStub.calledOnce);
          Assert(collection.insertStub.calledWith({
            userId: userId,
            date: Sinon.match.date,
            isAuthenticationSuccessful: true
          }));
          return BluebirdPromise.resolve();
        });
    });

    function should_retrieve_latest_authentication_traces(count: number) {
      factory.buildStub.returns(collection);
      collection.findStub.withArgs().returns(BluebirdPromise.resolve());

      const dataStore = new UserDataStore(factory);

      return dataStore.retrieveLatestAuthenticationTraces(userId, count)
        .then(function (doc: AuthenticationTraceDocument[]) {
          Assert(collection.findStub.calledOnce);
          Assert(collection.findStub.calledWith({
            userId: userId,
          }, { date: -1 }, count));
          return BluebirdPromise.resolve();
        });
    }

    it("should retrieve 3 latest failed authentication traces", function () {
      should_retrieve_latest_authentication_traces(3);
    });
  });


  describe("Identity validation collection", function () {
    it("should save a identity validation token", function () {
      factory.buildStub.returns(collection);
      collection.insertStub.returns(BluebirdPromise.resolve());

      const dataStore = new UserDataStore(factory);
      const maxAge = 400;
      const token = "TOKEN";
      const challenge = "CHALLENGE";

      return dataStore.produceIdentityValidationToken(userId, token, challenge, maxAge)
        .then(function (doc) {
          Assert(collection.insertStub.calledOnce);
          Assert(collection.insertStub.calledWith({
            userId: userId,
            token: token,
            challenge: challenge,
            maxDate: Sinon.match.date
          }));
          return BluebirdPromise.resolve();
        });
    });

    it("should consume an identity token successfully", function () {
      factory.buildStub.returns(collection);

      MockDate.set(100);

      const token = "TOKEN";
      const challenge = "CHALLENGE";

      collection.findOneStub.withArgs().returns(BluebirdPromise.resolve({
        userId: "USER",
        token: token,
        challenge: challenge,
        maxDate: new Date()
      }));
      collection.removeStub.returns(BluebirdPromise.resolve());

      const dataStore = new UserDataStore(factory);

      MockDate.set(80);

      return dataStore.consumeIdentityValidationToken(token, challenge)
        .then(function (doc) {
          MockDate.reset();
          Assert(collection.findOneStub.calledOnce);
          Assert(collection.findOneStub.calledWith({
            token: token,
            challenge: challenge
          }));

          Assert(collection.removeStub.calledOnce);
          Assert(collection.removeStub.calledWith({
            token: token,
            challenge: challenge
          }));
          return BluebirdPromise.resolve();
        });
    });

    it("should consume an expired identity token", function () {
      factory.buildStub.returns(collection);

      MockDate.set(0);

      const token = "TOKEN";
      const challenge = "CHALLENGE";

      collection.findOneStub.withArgs().returns(BluebirdPromise.resolve({
        userId: "USER",
        token: token,
        challenge: challenge,
        maxDate: new Date()
      }));

      const dataStore = new UserDataStore(factory);

      MockDate.set(80000);

      return dataStore.consumeIdentityValidationToken(token, challenge)
        .then(function () { return BluebirdPromise.reject(new Error("should not be here")); })
        .catch(function () {
          MockDate.reset();
          Assert(collection.findOneStub.calledOnce);
          Assert(collection.findOneStub.calledWith({
            token: token,
            challenge: challenge
          }));
          return BluebirdPromise.resolve();
        });
    });
  });
  describe("Prefered 2FA method", function () {
    it("should save a prefered 2FA method", async function () {
      factory.buildStub.returns(collection);
      collection.insertStub.returns(BluebirdPromise.resolve());

      const dataStore = new UserDataStore(factory);

      await dataStore.savePrefered2FAMethod(userId, "totp")
      Assert(collection.updateStub.calledOnce);
      Assert(collection.updateStub.calledWith(
        {userId}, {userId, method: "totp"}, {upsert: true}));
    });

    it("should retrieve a prefered 2FA method", async function () {
      factory.buildStub.returns(collection);
      collection.findOneStub.returns(BluebirdPromise.resolve());

      const dataStore = new UserDataStore(factory);

      await dataStore.retrievePrefered2FAMethod(userId)
      Assert(collection.findOneStub.calledOnce);
      Assert(collection.findOneStub.calledWith({userId}));
    });
  });
});
