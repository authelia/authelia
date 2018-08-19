import {setDefaultTimeout, After, Before, BeforeAll, AfterAll} from "cucumber";
import fs = require("fs");
import BluebirdPromise = require("bluebird");
import ChildProcess = require("child_process");
import { UserDataStore } from "../../../server/src/lib/storage/UserDataStore";
import { CollectionFactoryFactory } from "../../../server/src/lib/storage/CollectionFactoryFactory";
import { IMongoClient } from "../../../server/src/lib/connectors/mongo/IMongoClient";
import { TotpHandler } from "../../../server/src/lib/authentication/totp/TotpHandler";
import Speakeasy = require("speakeasy");
import Request = require("request-promise");
import { TOTPSecret } from "../../../server/types/TOTPSecret";
import Environment = require("../../environment");
import { MongoClient } from "../../../server/src/lib/connectors/mongo/MongoClient";
import { GlobalLogger } from "../../../server/src/lib/logging/GlobalLogger";
import { GlobalLoggerStub } from "../../../server/src/lib/logging/GlobalLoggerStub.spec";

setDefaultTimeout(30 * 1000);

const exec = BluebirdPromise.promisify<any, any>(ChildProcess.exec);

const includes = [
  "docker-compose.yml",
  "example/compose/docker-compose.base.yml",
  "example/compose/mongo/docker-compose.yml",
  "example/compose/redis/docker-compose.yml",
  "example/compose/nginx/backend/docker-compose.yml",
  "example/compose/nginx/portal/docker-compose.yml",
  "example/compose/smtp/docker-compose.yml",
  "example/compose/httpbin/docker-compose.yml",
  "example/compose/ldap/docker-compose.yml"
]

const environment = new Environment.Environment(includes);

BeforeAll(function() {
  return environment.setup(10000);
});

AfterAll(function() {
  return environment.cleanup()
});

Before(function () {
  this.jar = Request.jar();
})

After(function () {
  return this.driver.quit();
});

function createRegulationConfiguration(): BluebirdPromise<void> {
  return exec("\
  cat config.template.yml | \
  sed 's/find_time: [0-9]\\+/find_time: 15/' | \
  sed 's/ban_time: [0-9]\\+/ban_time: 4/' > config.test.yml \
  ");
}

function createInactivityConfiguration(): BluebirdPromise<void> {
  return exec("\
  cat config.template.yml | \
  sed 's/expiration: [0-9]\\+/expiration: 10000/' | \
  sed 's/inactivity: [0-9]\\+/inactivity: 5000/' > config.test.yml \
  ");
}

function createSingleFactorConfiguration(): BluebirdPromise<void> {
  return exec("\
  cat config.template.yml | \
  sed 's/default_method: two_factor/default_method: single_factor/' > config.test.yml \
  ");
}

function createCustomTotpIssuerConfiguration(): BluebirdPromise<void> {
  return exec("\
  cat config.template.yml | \
  sed 's/issuer: authelia.com/issuer: custom.com/' > config.test.yml \
  ");
}

function declareNeedsConfiguration(tag: string, cb: () => BluebirdPromise<void>) {
  Before({ tags: "@needs-" + tag + "-config", timeout: 20 * 1000 }, function () {
    return cb()
      .then(function () {
        return exec("./scripts/example-commit/dc-example.sh -f " +
          "./example/compose/authelia/docker-compose.test.yml up -d authelia &&" +
          " sleep 3");
      })
  });

  After({ tags: "@needs-" + tag + "-config", timeout: 20 * 1000 }, function () {
    return exec("rm config.test.yml")
      .then(function () {
        return exec("./scripts/example-commit/dc-example.sh up -d authelia && sleep 3");
      });
  });
}

declareNeedsConfiguration("regulation", createRegulationConfiguration);
declareNeedsConfiguration("inactivity", createInactivityConfiguration);
declareNeedsConfiguration("single_factor", createSingleFactorConfiguration);
declareNeedsConfiguration("totp_issuer", createCustomTotpIssuerConfiguration);

function registerUser(context: any, username: string) {
  let secret: TOTPSecret;
  const mongoClient = new MongoClient("mongodb://localhost:27017", "authelia", new GlobalLoggerStub());
  const collectionFactory = CollectionFactoryFactory.createMongo(mongoClient);
  const userDataStore = new UserDataStore(collectionFactory);

  const generator = new TotpHandler(Speakeasy);
  secret = generator.generate("user", "authelia.com");
  return userDataStore.saveTOTPSecret(username, secret)
    .then(function () {
      context.totpSecrets["REGISTERED"] = secret.base32;
      return mongoClient.close();
    });
}

function declareNeedRegisteredUserHooks(username: string) {
  Before({ tags: "@need-registered-user-" + username, timeout: 15 * 1000 }, function () {
    return registerUser(this, username);
  });

  After({ tags: "@need-registered-user-" + username, timeout: 15 * 1000 }, function () {
    this.totpSecrets["REGISTERED"] = undefined;
    return BluebirdPromise.resolve();
  });
}

function needAuthenticatedUser(context: any, username: string): BluebirdPromise<void> {
  return context.visit("https://login.example.com:8080/logout")
    .then(function () {
      return context.visit("https://login.example.com:8080/");
    })
    .then(function () {
      return registerUser(context, username);
    })
    .then(function () {
      return context.loginWithUserPassword(username, "password");
    })
    .then(function () {
      return context.useTotpTokenHandle("REGISTERED");
    })
    .then(function () {
      return context.clickOnButton("Sign in");
    });
}

function declareNeedAuthenticatedUserHooks(username: string) {
  Before({ tags: "@need-authenticated-user-" + username, timeout: 15 * 1000 }, function () {
    return needAuthenticatedUser(this, username);
  });

  After({ tags: "@need-authenticated-user-" + username, timeout: 15 * 1000 }, function () {
    this.totpSecrets["REGISTERED"] = undefined;
    return BluebirdPromise.resolve();
  });
}

function declareHooksForUser(username: string) {
  declareNeedRegisteredUserHooks(username);
  declareNeedAuthenticatedUserHooks(username);
}

const users = ["harry", "john", "bob", "blackhat"];
users.forEach(declareHooksForUser);
