
import AuthenticationRegulator from "../../src/lib/AuthenticationRegulator";
import UserDataStore from "../../src/lib/UserDataStore";
import * as MockDate from "mockdate";

const exceptions = require("../../src/lib/exceptions");

describe("test authentication regulator", function() {
  it("should mark 2 authentication and regulate (resolve)", function() {
    const options = {
      inMemoryOnly: true
    };
    const data_store = new UserDataStore(options);
    const regulator = new AuthenticationRegulator(data_store, 10);
    const user = "user";

    return regulator.mark(user, false)
    .then(function() {
      return regulator.mark(user, true);
    })
    .then(function() {
      return regulator.regulate(user);
    });
  });

  it("should mark 3 authentications and regulate (reject)", function(done) {
    const options = {
      inMemoryOnly: true
    };
    const data_store = new UserDataStore(options);
    const regulator = new AuthenticationRegulator(data_store, 10);
    const user = "user";

    regulator.mark(user, false)
    .then(function() {
      return regulator.mark(user, false);
    })
    .then(function() {
      return regulator.mark(user, false);
    })
    .then(function() {
      return regulator.regulate(user);
    })
    .catch(exceptions.AuthenticationRegulationError, function() {
      done();
    });
  });

  it("should mark 3 authentications and regulate (resolve)", function(done) {
    const options = {
      inMemoryOnly: true
    };
    const data_store = new UserDataStore(options);
    const regulator = new AuthenticationRegulator(data_store, 10);
    const user = "user";

    MockDate.set("1/2/2000 00:00:00");
    regulator.mark(user, false)
    .then(function() {
      MockDate.set("1/2/2000 00:00:15");
      return regulator.mark(user, false);
    })
    .then(function() {
      return regulator.mark(user, false);
    })
    .then(function() {
      return regulator.regulate(user);
    })
    .then(function() {
      done();
    });
  });
});
