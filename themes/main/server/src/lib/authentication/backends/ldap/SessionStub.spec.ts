import Bluebird = require("bluebird");
import Sinon = require("sinon");

import { ISession } from "./ISession";

export class SessionStub implements ISession {
  openStub: Sinon.SinonStub;
  closeStub: Sinon.SinonStub;
  searchUserDnStub: Sinon.SinonStub;
  searchEmailsStub: Sinon.SinonStub;
  searchGroupsStub: Sinon.SinonStub;
  modifyPasswordStub: Sinon.SinonStub;

  constructor() {
    this.openStub = Sinon.stub();
    this.closeStub = Sinon.stub();
    this.searchUserDnStub = Sinon.stub();
    this.searchEmailsStub = Sinon.stub();
    this.searchGroupsStub = Sinon.stub();
    this.modifyPasswordStub = Sinon.stub();
  }

  open(): Bluebird<void> {
    return this.openStub();
  }

  close(): Bluebird<void> {
    return this.closeStub();
  }

  searchUserDn(username: string): Bluebird<string> {
    return this.searchUserDnStub(username);
  }

  searchEmails(username: string): Bluebird<string[]> {
    return this.searchEmailsStub(username);
  }

  searchGroups(username: string): Bluebird<string[]> {
    return this.searchGroupsStub(username);
  }

  modifyPassword(username: string, newPassword: string): Bluebird<void> {
    return this.modifyPasswordStub(username, newPassword);
  }
}