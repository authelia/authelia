import BluebirdPromise = require("bluebird");
import { ISession } from "./ISession";
import { Sanitizer } from "./Sanitizer";
import { Winston } from "../../../../../types/Dependencies";

const SPECIAL_CHAR_USED_MESSAGE = "Special character used in LDAP query.";


export class SafeSession implements ISession {
  private sesion: ISession;
  private logger: Winston;

  constructor(sesion: ISession, logger: Winston) {
    this.sesion = sesion;
    this.logger = logger;
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
      this.logger.error("Error with input " + username + ". Cause:" + e);
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }

  searchUserDn(username: string): BluebirdPromise<string> {
    try {
      const sanitizedUsername = Sanitizer.sanitize(username);
      return this.sesion.searchUserDn(sanitizedUsername);
    }
    catch (e) {
      this.logger.error("Error with input " + username + ". Cause:" + e);
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }

  searchEmails(username: string): BluebirdPromise<string[]> {
    try {
      const sanitizedUsername = Sanitizer.sanitize(username);
      return this.sesion.searchEmails(sanitizedUsername);
    }
    catch (e) {
      this.logger.error("Error with input " + username + ". Cause:" + e);
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }

  modifyPassword(username: string, newPassword: string): BluebirdPromise<void> {
    try {
      const sanitizedUsername = Sanitizer.sanitize(username);
      return this.sesion.modifyPassword(sanitizedUsername, newPassword);
    }
    catch (e) {
      this.logger.error("Error with input " + username + ". Cause:" + e);
      return BluebirdPromise.reject(new Error(SPECIAL_CHAR_USED_MESSAGE));
    }
  }
}
