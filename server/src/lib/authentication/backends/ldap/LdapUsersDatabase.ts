import Bluebird = require("bluebird");
import { IUsersDatabase } from "../IUsersDatabase";
import { ISessionFactory } from "./ISessionFactory";
import { LdapConfiguration } from "../../../configuration/schema/LdapConfiguration";
import { ISession } from "./ISession";
import { GroupsAndEmails } from "../GroupsAndEmails";
import Exceptions = require("../../../Exceptions");
import AuthenticationError from "../../AuthenticationError";

type SessionCallback<T> = (session: ISession) => Bluebird<T>;

export class LdapUsersDatabase implements IUsersDatabase {
  private sessionFactory: ISessionFactory;
  private configuration: LdapConfiguration;

  constructor(
    sessionFactory: ISessionFactory,
    configuration: LdapConfiguration) {
    this.sessionFactory = sessionFactory;
    this.configuration = configuration;
  }

  private withSession<T>(
    username: string,
    password: string,
    cb: SessionCallback<T>): Bluebird<T> {
    const session = this.sessionFactory.create(username, password);
    return session.open()
      .then(() => cb(session))
      .finally(() => session.close());
  }

  checkUserPassword(username: string, password: string): Bluebird<GroupsAndEmails> {
    const that = this;
    function verifyUserPassword(userDN: string) {
      return that.withSession<void>(
        userDN,
        password,
        (session) => Bluebird.resolve()
      );
    }

    function getInfo(session: ISession) {
        return Bluebird.join(
          session.searchGroups(username),
          session.searchEmails(username)
        )
        .spread((groups: string[], emails: string[]) => {
          return { groups: groups, emails: emails };
        });
    }

    return that.withSession(
      that.configuration.user,
      that.configuration.password,
      (session) => {
        return session.searchUserDn(username)
          .then(verifyUserPassword)
          .then(() => getInfo(session));
      })
      .catch((err) =>
        Bluebird.reject(new AuthenticationError(err.message)));
  }

  getEmails(username: string): Bluebird<string[]> {
    const that = this;
    return that.withSession(
      that.configuration.user,
      that.configuration.password,
      (session) => {
        return session.searchEmails(username);
      }
    )
    .catch((err) =>
      Bluebird.reject(new Exceptions.LdapError("Failed during email retrieval: " + err.message))
    );
  }

  getGroups(username: string): Bluebird<string[]> {
    const that = this;
    return that.withSession(
      that.configuration.user,
      that.configuration.password,
      (session) => {
        return session.searchGroups(username);
      }
    )
    .catch((err) =>
      Bluebird.reject(new Exceptions.LdapError("Failed during email retrieval: " + err.message))
    );
  }

  updatePassword(username: string, newPassword: string): Bluebird<void> {
    const that = this;
    return that.withSession(
      that.configuration.user,
      that.configuration.password,
      (session) => {
        return session.modifyPassword(username, newPassword);
      }
    )
    .catch(function (err: Error) {
      return Bluebird.reject(
        new Exceptions.LdapError(
          "Error while updating password: " + err.message));
    });
  }
}