
import * as BluebirdPromise from "bluebird";
import exceptions = require("../Exceptions");
import { IUserDataStore } from "../storage/IUserDataStore";
import { AuthenticationTraceDocument } from "../storage/AuthenticationTraceDocument";
import { IRegulator } from "./IRegulator";

export class Regulator implements IRegulator {
  private userDataStore: IUserDataStore;
  private banTime: number;
  private findTime: number;
  private maxRetries: number;

  constructor(userDataStore: any, maxRetries: number, findTime: number, banTime: number) {
    this.userDataStore = userDataStore;
    this.banTime = banTime;
    this.findTime = findTime;
    this.maxRetries = maxRetries;
  }

  // Mark authentication
  mark(userId: string, isAuthenticationSuccessful: boolean): BluebirdPromise<void> {
    return this.userDataStore.saveAuthenticationTrace(userId, isAuthenticationSuccessful);
  }

  regulate(userId: string): BluebirdPromise<void> {
    const that = this;

    if (that.maxRetries <= 0) return BluebirdPromise.resolve();

    return this.userDataStore.retrieveLatestAuthenticationTraces(userId, that.maxRetries)
      .then((docs: AuthenticationTraceDocument[]) => {
        // less than the max authorized number of authentication in time range, thus authorizing access
        if (docs.length < that.maxRetries) return BluebirdPromise.resolve();

        const numberOfFailedAuth = docs
          .map(function (d: AuthenticationTraceDocument) { return d.isAuthenticationSuccessful == false ? 1 : 0; })
          .reduce(function (acc, v) { return acc + v; }, 0);

        if (numberOfFailedAuth < this.maxRetries) return BluebirdPromise.resolve();

        const newestDocument = docs[0];
        const oldestDocument = docs[that.maxRetries - 1];

        const authenticationsTimeRangeInSeconds = (newestDocument.date.getTime() - oldestDocument.date.getTime()) / 1000;
        const tooManyAuthInTimelapse = (authenticationsTimeRangeInSeconds < this.findTime);
        const stillInBannedTimeRange = (new Date(new Date().getTime() - this.banTime * 1000) < newestDocument.date);

        if (tooManyAuthInTimelapse && stillInBannedTimeRange)
          throw new exceptions.AuthenticationRegulationError("Max number of authentication. Please retry in few minutes.");

        return BluebirdPromise.resolve();
      });
  }
}
