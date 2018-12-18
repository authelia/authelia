import Bluebird = require("bluebird");
import Sinon = require("sinon");

import { IUsersDatabase } from "./IUsersDatabase";
import { GroupsAndEmails } from "./GroupsAndEmails";

export class IUsersDatabaseStub implements IUsersDatabase {
  checkUserPasswordStub: Sinon.SinonStub;
  getEmailsStub: Sinon.SinonStub;
  getGroupsStub: Sinon.SinonStub;
  updatePasswordStub: Sinon.SinonStub;

  constructor() {
    this.checkUserPasswordStub = Sinon.stub();
    this.getEmailsStub = Sinon.stub();
    this.getGroupsStub = Sinon.stub();
    this.updatePasswordStub = Sinon.stub();
  }

  checkUserPassword(username: string, password: string): Bluebird<GroupsAndEmails> {
    return this.checkUserPasswordStub(username, password);
  }

  getEmails(username: string): Bluebird<string[]> {
    return this.getEmailsStub(username);
  }

  getGroups(username: string): Bluebird<string[]> {
    return this.getGroupsStub(username);
  }

  updatePassword(username: string, newPassword: string): Bluebird<void> {
    return this.updatePasswordStub(username, newPassword);
  }
}