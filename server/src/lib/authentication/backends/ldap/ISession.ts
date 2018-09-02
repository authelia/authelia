
import BluebirdPromise = require("bluebird");
import { UsersWithNetworkAddresses } from "../UsersWithNetworkAddresses";

export interface ISession {
  open(): BluebirdPromise<void>;
  close(): BluebirdPromise<void>;

  searchUserDn(username: string): BluebirdPromise<string>;
  searchEmails(username: string): BluebirdPromise<string[]>;
  searchGroups(username: string): BluebirdPromise<string[]>;
  searchWhitelist(): BluebirdPromise<UsersWithNetworkAddresses[]>;
  modifyPassword(username: string, newPassword: string): BluebirdPromise<void>;
}