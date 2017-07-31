
import * as BluebirdPromise from "bluebird";
import exceptions = require("./Exceptions");
import { UserDataStore } from "./storage/UserDataStore";
import { AuthenticationTraceDocument } from "./storage/AuthenticationTraceDocument";

const MAX_AUTHENTICATION_COUNT_IN_TIME_RANGE = 3;

export class AuthenticationRegulator {
  private userDataStore: UserDataStore;
  private lockTimeInSeconds: number;

  constructor(userDataStore: any, lockTimeInSeconds: number) {
    this.userDataStore = userDataStore;
    this.lockTimeInSeconds = lockTimeInSeconds;
  }

  // Mark authentication
  mark(userId: string, isAuthenticationSuccessful: boolean): BluebirdPromise<void> {
    return this.userDataStore.saveAuthenticationTrace(userId, isAuthenticationSuccessful);
  }

  regulate(userId: string): BluebirdPromise<void> {
    return this.userDataStore.retrieveLatestAuthenticationTraces(userId, false, 3)
      .then((docs: AuthenticationTraceDocument[]) => {
        if (docs.length < MAX_AUTHENTICATION_COUNT_IN_TIME_RANGE) {
          // less than the max authorized number of authentication in time range, thus authorizing access
          return BluebirdPromise.resolve();
        }

        const oldestDocument = docs[MAX_AUTHENTICATION_COUNT_IN_TIME_RANGE - 1];
        const noLockMinDate = new Date(new Date().getTime() - this.lockTimeInSeconds * 1000);
        if (oldestDocument.date > noLockMinDate) {
          throw new exceptions.AuthenticationRegulationError("Max number of authentication. Please retry in few minutes.");
        }

        return BluebirdPromise.resolve();
      });
  }
}
