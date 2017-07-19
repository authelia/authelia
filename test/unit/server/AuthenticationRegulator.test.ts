
import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");

import { AuthenticationRegulator } from "../../../src/server/lib/AuthenticationRegulator";
import { UserDataStore } from "../../../src/server/lib/storage/UserDataStore";
import MockDate = require("mockdate");
import exceptions = require("../../../src/server/lib/Exceptions");
import { CollectionStub } from "./mocks/storage/CollectionStub";
import { CollectionFactoryStub } from "./mocks/storage/CollectionFactoryStub";

describe("test authentication regulator", function () {
  let collectionFactory: CollectionFactoryStub;
  let collection: CollectionStub;

  beforeEach(function () {
    collectionFactory = new CollectionFactoryStub();
    collection = new CollectionStub();

    collectionFactory.buildStub.returns(collection);
  });

  it("should mark 2 authentication and regulate", function () {
    const user = "USER";

    collection.insertStub.returns(BluebirdPromise.resolve());
    collection.findStub.returns(BluebirdPromise.resolve([{
      userId: user,
      date: new Date(),
      isAuthenticationSuccessful: false
    }, {
      userId: user,
      date: new Date(),
      isAuthenticationSuccessful: true
    }]));

    const dataStore = new UserDataStore(collectionFactory);
    const regulator = new AuthenticationRegulator(dataStore, 10);

    return regulator.mark(user, false)
      .then(function () {
        return regulator.mark(user, true);
      })
      .then(function () {
        return regulator.regulate(user);
      });
  });

  it("should mark 3 authentications and regulate (reject)", function (done) {
    const user = "USER";
    collection.insertStub.returns(BluebirdPromise.resolve());
    collection.findStub.returns(BluebirdPromise.resolve([{
      userId: user,
      date: new Date(),
      isAuthenticationSuccessful: false
    }, {
      userId: user,
      date: new Date(),
      isAuthenticationSuccessful: false
    }, {
      userId: user,
      date: new Date(),
      isAuthenticationSuccessful: false
    }]));

    const dataStore = new UserDataStore(collectionFactory);
    const regulator = new AuthenticationRegulator(dataStore, 10);

    regulator.mark(user, false)
      .then(function () {
        return regulator.mark(user, false);
      })
      .then(function () {
        return regulator.mark(user, false);
      })
      .then(function () {
        return regulator.regulate(user);
      })
      .catch(exceptions.AuthenticationRegulationError, function () {
        done();
      });
  });

  it("should mark 3 authentications separated by a lot of time and allow access to user", function (done) {
    const user = "USER";
    collection.insertStub.returns(BluebirdPromise.resolve());
    collection.findStub.returns(BluebirdPromise.resolve([{
      userId: user,
      date: new Date("1/2/2000 06:00:15"),
      isAuthenticationSuccessful: false
    }, {
      userId: user,
      date: new Date("1/2/2000 00:00:15"),
      isAuthenticationSuccessful: false
    }, {
      userId: user,
      date: new Date("1/2/2000 00:00:00"),
      isAuthenticationSuccessful: false
    }]));
    const data_store = new UserDataStore(collectionFactory);
    const regulator = new AuthenticationRegulator(data_store, 10);

    MockDate.set("1/2/2000 00:00:00");
    regulator.mark(user, false)
      .then(function () {
        MockDate.set("1/2/2000 00:00:15");
        return regulator.mark(user, false);
      })
      .then(function () {
        MockDate.set("1/2/2000 06:00:15");
        return regulator.mark(user, false);
      })
      .then(function () {
        return regulator.regulate(user);
      })
      .then(function () {
        done();
      });
  });
});