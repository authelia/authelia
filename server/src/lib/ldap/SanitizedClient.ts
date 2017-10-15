import BluebirdPromise = require("bluebird");
import { IClient, GroupsAndEmails } from "./IClient";
import { Client } from "./Client";
import { InputsSanitizer } from "./InputsSanitizer";

const SPECIAL_CHAR_USED_MESSAGE = "Special character used in LDAP query.";


export class SanitizedClient implements IClient {
  private client: IClient;

  constructor(client: IClient) {
    this.client = client;
  }

  open(): BluebirdPromise<void> {
    return this.client.open();
  }

  close(): BluebirdPromise<void> {
    return this.client.close();
  }

  searchGroups(username: string): BluebirdPromise<string[]> {
    try {
      const sanitizedUsername = InputsSanitizer.sanitize(username);
      return this.client.searchGroups(sanitizedUsername);
    }
    catch (e) {
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }

  searchUserDn(username: string): BluebirdPromise<string> {
    try {
      const sanitizedUsername = InputsSanitizer.sanitize(username);
      return this.client.searchUserDn(sanitizedUsername);
    }
    catch (e) {
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }

  searchEmails(username: string): BluebirdPromise<string[]> {
    try {
      const sanitizedUsername = InputsSanitizer.sanitize(username);
      return this.client.searchEmails(sanitizedUsername);
    }
    catch (e) {
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }

  modifyPassword(username: string, newPassword: string): BluebirdPromise<void> {
    try {
      const sanitizedUsername = InputsSanitizer.sanitize(username);
      return this.client.modifyPassword(sanitizedUsername, newPassword);
    }
    catch (e) {
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }
}
