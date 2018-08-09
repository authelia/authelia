
import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");

import { Regulator } from "./Regulator";
import MockDate = require("mockdate");
import exceptions = require("../Exceptions");
import { UserDataStoreStub } from "../storage/UserDataStoreStub.spec";

describe("regulation/Regulator", function () {
  const USER1 = "USER1";
  const USER2 = "USER2";
  let userDataStoreStub: UserDataStoreStub;

  beforeEach(function () {
    userDataStoreStub = new UserDataStoreStub();
    const dataStore: { [userId: string]: { userId: string, date: Date, isAuthenticationSuccessful: boolean }[] } = {
      [USER1]: [],
      [USER2]: []
    };

    userDataStoreStub.saveAuthenticationTraceStub.callsFake(function (userId, isAuthenticationSuccessful) {
      dataStore[userId].unshift({
        userId: userId,
        date: new Date(),
        isAuthenticationSuccessful: isAuthenticationSuccessful,
      });
      return BluebirdPromise.resolve();
    });

    userDataStoreStub.retrieveLatestAuthenticationTracesStub.callsFake(function (userId, count) {
      const ret = (dataStore[userId].length <= count) ? dataStore[userId] : dataStore[userId].slice(0, 3);
      return BluebirdPromise.resolve(ret);
    });
  });

  afterEach(function () {
    MockDate.reset();
  });

  function markAuthenticationAt(regulator: Regulator, user: string, time: string, success: boolean) {
    MockDate.set(time);
    return regulator.mark(user, success);
  }

  it("should mark 2 authentication and regulate (accept)", function () {
    const regulator = new Regulator(userDataStoreStub, 3, 10, 10);

    return regulator.mark(USER1, false)
      .then(function () {
        return regulator.mark(USER1, true);
      })
      .then(function () {
        return regulator.regulate(USER1);
      });
  });

  it("should mark 3 authentications and regulate (reject)", function () {
    const regulator = new Regulator(userDataStoreStub, 3, 10, 10);

    return regulator.mark(USER1, false)
      .then(function () {
        return regulator.mark(USER1, false);
      })
      .then(function () {
        return regulator.mark(USER1, false);
      })
      .then(function () {
        return regulator.regulate(USER1);
      })
      .then(function () { return BluebirdPromise.reject(new Error("should not be here!")); })
      .catch(exceptions.AuthenticationRegulationError, function () {
        return BluebirdPromise.resolve();
      });
  });

  it("should mark 1 failed, 1 successful and 1 failed authentications within minimum time and regulate (accept)", function () {
    const regulator = new Regulator(userDataStoreStub, 3, 60, 30);

    return markAuthenticationAt(regulator, USER1, "1/2/2000 00:00:00", false)
      .then(function () {
        return markAuthenticationAt(regulator, USER1, "1/2/2000 00:00:10", true);
      })
      .then(function () {
        return markAuthenticationAt(regulator, USER1, "1/2/2000 00:00:20", false);
      })
      .then(function () {
        return regulator.regulate(USER1);
      })
      .then(function () {
        return markAuthenticationAt(regulator, USER1, "1/2/2000 00:00:30", false);
      })
      .then(function () {
        return regulator.regulate(USER1);
      })
      .then(function () {
        return markAuthenticationAt(regulator, USER1, "1/2/2000 00:00:39", false);
      })
      .then(function () {
        return regulator.regulate(USER1);
      })
      .then(function () {
        return BluebirdPromise.reject(new Error("should not be here!"));
      },
      function () {
        return BluebirdPromise.resolve();
      });
  });

  it("should regulate user if number of failures is greater than 3 in allowed time lapse", function () {
    function markAuthentications(regulator: Regulator, user: string) {
      return markAuthenticationAt(regulator, user, "1/2/2000 00:00:00", false)
        .then(function () {
          return markAuthenticationAt(regulator, user, "1/2/2000 00:00:45", false);
        })
        .then(function () {
          return markAuthenticationAt(regulator, user, "1/2/2000 00:01:05", false);
        })
        .then(function () {
          return regulator.regulate(user);
        });
    }

    const regulator1 = new Regulator(userDataStoreStub, 3, 60, 60);
    const regulator2 = new Regulator(userDataStoreStub, 3, 2 * 60, 60);

    const p1 = markAuthentications(regulator1, USER1);
    const p2 = markAuthentications(regulator2, USER2);

    return BluebirdPromise.join(p1, p2)
      .then(function () {
        return BluebirdPromise.reject(new Error("should not be here..."));
      }, function () {
        Assert(p1.isFulfilled());
        Assert(p2.isRejected());
      });
  });

  it("should user wait after regulation to authenticate again", function () {
    function markAuthentications(regulator: Regulator, user: string) {
      return markAuthenticationAt(regulator, user, "1/2/2000 00:00:00", false)
        .then(function () {
          return markAuthenticationAt(regulator, user, "1/2/2000 00:00:10", false);
        })
        .then(function () {
          return markAuthenticationAt(regulator, user, "1/2/2000 00:00:15", false);
        })
        .then(function () {
          return markAuthenticationAt(regulator, user, "1/2/2000 00:00:25", false);
        })
        .then(function () {
          MockDate.set("1/2/2000 00:00:54");
          return regulator.regulate(user);
        })
        .then(function () {
          return BluebirdPromise.reject(new Error("should fail at this time"));
        }, function () {
          MockDate.set("1/2/2000 00:00:56");
          return regulator.regulate(user);
        });
    }

    const regulator = new Regulator(userDataStoreStub, 4, 30, 30);
    return markAuthentications(regulator, USER1);
  });

  it("should disable regulation when max_retries is set to 0", function () {
    const maxRetries = 0;
    const regulator = new Regulator(userDataStoreStub, maxRetries, 60, 30);
    return markAuthenticationAt(regulator, USER1, "1/2/2000 00:00:00", false)
      .then(function () {
        return markAuthenticationAt(regulator, USER1, "1/2/2000 00:00:10", false);
      })
      .then(function () {
        return markAuthenticationAt(regulator, USER1, "1/2/2000 00:00:15", false);
      })
      .then(function () {
        return markAuthenticationAt(regulator, USER1, "1/2/2000 00:00:25", false);
      })
      .then(function () {
        MockDate.set("1/2/2000 00:00:26");
        return regulator.regulate(USER1);
      });
  });
});