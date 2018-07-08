
import BluebirdPromise = require("bluebird");
import { IClient, GroupsAndEmails } from "./IClient";
import Sinon = require("sinon");

export class ClientStub implements IClient {
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

  open(): BluebirdPromise<void> {
    return this.openStub();
  }

  close(): BluebirdPromise<void> {
    return this.closeStub();
  }

  searchUserDn(username: string): BluebirdPromise<string> {
    return this.searchUserDnStub(username);
  }

  searchEmails(username: string): BluebirdPromise<string[]> {
    return this.searchEmailsStub(username);
  }

  searchGroups(username: string): BluebirdPromise<string[]> {
    return this.searchGroupsStub(username);
  }

  modifyPassword(username: string, newPassword: string): BluebirdPromise<void> {
    return this.modifyPasswordStub(username, newPassword);
  }
}