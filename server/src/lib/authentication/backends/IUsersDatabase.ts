import Bluebird = require("bluebird");

import { GroupsAndEmails } from "./GroupsAndEmails";
import { UserAndNetworkAddresses } from "./UserAndNetworkAddresses";

export interface IUsersDatabase {
  checkUserPassword(username: string, password: string): Bluebird<GroupsAndEmails>;
  getEmails(username: string): Bluebird<string[]>;
  getGroups(username: string): Bluebird<string[]>;
  getUserAndNetworkAddresses(): Bluebird<UserAndNetworkAddresses[]>;
  updatePassword(username: string, newPassword: string): Bluebird<void>;
}