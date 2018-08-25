import Bluebird = require("bluebird");

import { GroupsAndEmails } from "./GroupsAndEmails";

export interface IUsersDatabase {
  checkUserPassword(username: string, password: string): Bluebird<GroupsAndEmails>;
  getEmails(username: string): Bluebird<string[]>;
  getGroups(username: string): Bluebird<string[]>;
  updatePassword(username: string, newPassword: string): Bluebird<void>;
}