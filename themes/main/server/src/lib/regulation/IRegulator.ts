import BluebirdPromise = require("bluebird");

export interface IRegulator {
  mark(userId: string, isAuthenticationSuccessful: boolean): BluebirdPromise<void>;
  regulate(userId: string): BluebirdPromise<void>;
}