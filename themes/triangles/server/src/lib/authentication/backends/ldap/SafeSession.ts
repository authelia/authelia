import BluebirdPromise = require("bluebird");
import { ISession } from "./ISession";
import { Sanitizer } from "./Sanitizer";

const SPECIAL_CHAR_USED_MESSAGE = "Special character used in LDAP query.";


export class SafeSession implements ISession {
  private sesion: ISession;

  constructor(sesion: ISession) {
    this.sesion = sesion;
  }

  open(): BluebirdPromise<void> {
    return this.sesion.open();
  }

  close(): BluebirdPromise<void> {
    return this.sesion.close();
  }

  searchGroups(username: string): BluebirdPromise<string[]> {
    try {
      const sanitizedUsername = Sanitizer.sanitize(username);
      return this.sesion.searchGroups(sanitizedUsername);
    }
    catch (e) {
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }

  searchUserDn(username: string): BluebirdPromise<string> {
    try {
      const sanitizedUsername = Sanitizer.sanitize(username);
      return this.sesion.searchUserDn(sanitizedUsername);
    }
    catch (e) {
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }

  searchEmails(username: string): BluebirdPromise<string[]> {
    try {
      const sanitizedUsername = Sanitizer.sanitize(username);
      return this.sesion.searchEmails(sanitizedUsername);
    }
    catch (e) {
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }

  modifyPassword(username: string, newPassword: string): BluebirdPromise<void> {
    try {
      const sanitizedUsername = Sanitizer.sanitize(username);
      return this.sesion.modifyPassword(sanitizedUsername, newPassword);
    }
    catch (e) {
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }
}
