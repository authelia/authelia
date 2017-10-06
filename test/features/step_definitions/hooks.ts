import Cucumber = require("cucumber");
import fs = require("fs");
import BluebirdPromise = require("bluebird");
import ChildProcess = require("child_process");
import { UserDataStore } from "../../../server/src/lib/storage/UserDataStore";
import { CollectionFactoryFactory } from "../../../server/src/lib/storage/CollectionFactoryFactory";
import { MongoConnector } from "../../../server/src/lib/connectors/mongo/MongoConnector";
import { IMongoClient } from "../../../server/src/lib/connectors/mongo/IMongoClient";
import { TOTPGenerator } from "../../../server/src/lib/TOTPGenerator";
import Speakeasy = require("speakeasy");

Cucumber.defineSupportCode(function ({ setDefaultTimeout }) {
  setDefaultTimeout(20 * 1000);
});

Cucumber.defineSupportCode(function ({ After, Before }) {
  const exec = BluebirdPromise.promisify(ChildProcess.exec);

  After(function () {
    return this.driver.quit();
  });

  Before({ tags: "@needs-test-config", timeout: 20 * 1000 }, function () {
    return exec("./scripts/example-commit/dc-example.sh -f docker-compose.test.yml up -d authelia && sleep 2");
  });

  After({ tags: "@needs-test-config", timeout: 20 * 1000 }, function () {
    return exec("./scripts/example-commit/dc-example.sh up -d authelia && sleep 2");
  });

  function registerUser(context: any, username: string) {
    let secret: Speakeasy.Key;
    const mongoConnector = new MongoConnector("mongodb://localhost:27017/authelia");
    return mongoConnector.connect()
      .then(function (mongoClient: IMongoClient) {
        const collectionFactory = CollectionFactoryFactory.createMongo(mongoClient);
        const userDataStore = new UserDataStore(collectionFactory);

        const generator = new TOTPGenerator(Speakeasy);
        secret = generator.generate();
        return userDataStore.saveTOTPSecret(username, secret);
      })
      .then(function () {
        context.totpSecrets["REGISTERED"] = secret.base32;
      });
  }

  function declareNeedRegisteredUserHooks(username: string) {
    Before({ tags: "@need-registered-user-" + username, timeout: 15 * 1000 }, function () {
      return registerUser(this, username);
    });

    After({ tags: "@need-registered-user-" + username, timeout: 15 * 1000 }, function () {
      this.totpSecrets["REGISTERED"] = undefined;
    });
  }

  function needAuthenticatedUser(context: any, username: string): BluebirdPromise<void> {
    return context.visit("https://auth.test.local:8080/")
      .then(function () {
        return registerUser(context, username);
      })
      .then(function () {
        return context.loginWithUserPassword(username, "password");
      })
      .then(function () {
        return context.useTotpTokenHandle("REGISTERED");
      })
      .then(function() {
        return context.clickOnButton("TOTP");
      });
  }

  function declareNeedAuthenticatedUserHooks(username: string) {
    Before({ tags: "@need-authenticated-user-" + username, timeout: 15 * 1000 }, function () {
      return needAuthenticatedUser(this, username);
    });

    After({ tags: "@need-authenticated-user-" + username, timeout: 15 * 1000 }, function () {
      this.totpSecrets["REGISTERED"] = undefined;
    });
  }

  function declareHooksForUser(username: string) {
    declareNeedRegisteredUserHooks(username);
    declareNeedAuthenticatedUserHooks(username);
  }

  const users = ["harry", "john", "bob", "blackhat"];
  users.forEach(declareHooksForUser);
});