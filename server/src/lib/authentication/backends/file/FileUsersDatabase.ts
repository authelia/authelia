import Bluebird = require("bluebird");
import Fs = require("fs");
import Yaml = require("yamljs");

import { FileUsersDatabaseConfiguration }
  from "../../../configuration/schema/FileUsersDatabaseConfiguration";
import { GroupsAndEmails } from "../GroupsAndEmails";
import { IUsersDatabase } from "../IUsersDatabase";
import { HashGenerator } from "../../../utils/HashGenerator";
import { ReadWriteQueue } from "./ReadWriteQueue";
import AuthenticationError from "../../AuthenticationError";

const loadAsync = Bluebird.promisify(Yaml.load);

export class FileUsersDatabase implements IUsersDatabase {
  private configuration: FileUsersDatabaseConfiguration;
  private queue: ReadWriteQueue;

  constructor(configuration: FileUsersDatabaseConfiguration) {
    this.configuration = configuration;
    this.queue = new ReadWriteQueue(this.configuration.path);
  }

  /**
   * Read database from file.
   * It enqueues the read task so that it is scheduled
   * between other reads and writes.
   */
  private readDatabase(): Bluebird<any> {
    return new Bluebird<string>((resolve, reject) => {
      this.queue.read((err: Error, data: string) => {
        if (err) {
          reject(err);
          return;
        }
        resolve(data);
        this.queue.next();
      });
    })
      .then((content) => {
        const database = Yaml.parse(content);
        if (!database) {
          return Bluebird.reject(new Error("Unable to parse YAML file."));
        }
        return Bluebird.resolve(database);
      });
  }

  /**
   * Checks the user exists in the database.
   */
  private checkUserExists(
    database: any,
    username: string)
    : Bluebird<void> {
    if (!(username in database.users)) {
      return Bluebird.reject(
        new Error(`User ${username} does not exist in database.`));
    }
    return Bluebird.resolve();
  }

  /**
   * Check the password of a given user.
   */
  private checkPassword(
    database: any,
    username: string,
    password: string)
    : Bluebird<void> {
    const storedHash: string = database.users[username].password;
    const matches = storedHash.match(/rounds=([0-9]+)\$([a-zA-z0-9./]+)\$/);
    if (!(matches && matches.length == 3)) {
      return Bluebird.reject(new Error("Unable to detect the hash salt and rounds. " +
        "Make sure the password is hashed with SSHA512."));
    }

    const rounds: number = parseInt(matches[1]);
    const salt = matches[2];

    return HashGenerator.ssha512(password, rounds, salt)
      .then((hash: string) => {
        if (hash !== storedHash) {
          return Bluebird.reject(new AuthenticationError("Wrong username/password."));
        }
        return Bluebird.resolve();
      });
  }

  /**
   * Retrieve email addresses of a given user.
   */
  private retrieveEmails(
    database: any,
    username: string)
    : Bluebird<string[]> {
    if (!("email" in database.users[username])) {
      return Bluebird.reject(
        new Error(`User ${username} has no email address.`));
    }
    return Bluebird.resolve(
      [database.users[username].email]);
  }

  private retrieveGroups(
    database: any,
    username: string)
    : Bluebird<string[]> {
    if (!("groups" in database.users[username])) {
      return Bluebird.resolve([]);
    }
    return Bluebird.resolve(
      database.users[username].groups);
  }

  private replacePassword(
    database: any,
    username: string,
    newPassword: string)
    : Bluebird<void> {
    const that = this;
    return HashGenerator.ssha512(newPassword)
      .then((hash) => {
        database.users[username].password = hash;
        const str = Yaml.stringify(database, 4, 2);
        return Bluebird.resolve(str);
      })
      .then((content: string) => {
        return new Bluebird((resolve, reject) => {
          that.queue.write(content, (err) => {
            if (err) {
              return reject(err);
            }
            resolve();
            that.queue.next();
          });
        });
      });
  }

  checkUserPassword(
    username: string,
    password: string)
    : Bluebird<GroupsAndEmails> {
    return this.readDatabase()
      .then((database) => {
        return this.checkUserExists(database, username)
          .then(() => this.checkPassword(database, username, password))
          .then(() => {
            return Bluebird.join(
              this.retrieveEmails(database, username),
              this.retrieveGroups(database, username)
            ).spread((emails: string[], groups: string[]) => {
              return { emails: emails, groups: groups };
            });
          });
      });
  }

  getEmails(username: string): Bluebird<string[]> {
    return this.readDatabase()
      .then((database) => {
        return this.checkUserExists(database, username)
          .then(() => this.retrieveEmails(database, username));
      });
  }

  getGroups(username: string): Bluebird<string[]> {
    return this.readDatabase()
      .then((database) => {
        return this.checkUserExists(database, username)
          .then(() => this.retrieveGroups(database, username));
      });
  }

  updatePassword(username: string, newPassword: string): Bluebird<void> {
    return this.readDatabase()
      .then((database) => {
        return this.checkUserExists(database, username)
          .then(() => this.replacePassword(database, username, newPassword));
      });
  }
}