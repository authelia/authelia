import Bluebird = require("bluebird");

import { GroupsAndEmails } from "./GroupsAndEmails";
import { UsersWithNetworkAddresses } from "./UsersWithNetworkAddresses";

export interface IUsersDatabase {
  checkUserPassword(username: string, password: string): Bluebird<GroupsAndEmails>;
  getEmails(username: string): Bluebird<string[]>;
  getGroups(username: string): Bluebird<string[]>;
  getUsersWithNetworkAddresses(): Bluebird<UsersWithNetworkAddresses[]>;
  updatePassword(username: string, newPassword: string): Bluebird<void>;
}