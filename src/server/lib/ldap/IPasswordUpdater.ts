import BluebirdPromise = require("bluebird");

export interface IPasswordUpdater {
  updatePassword(username: string, newPassword: string): BluebirdPromise<void>;
}