
import BluebirdPromise = require("bluebird");

export interface GroupsAndEmails {
  groups: string[];
  emails: string[];
}

export interface IClient {
  open(): BluebirdPromise<void>;
  close(): BluebirdPromise<void>;
  searchUserDn(username: string): BluebirdPromise<string>;
  searchEmails(username: string): BluebirdPromise<string[]>;
  searchEmailsAndGroups(username: string): BluebirdPromise<GroupsAndEmails>;
  modifyPassword(username: string, newPassword: string): BluebirdPromise<void>;
}