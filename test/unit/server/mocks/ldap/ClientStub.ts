
import BluebirdPromise = require("bluebird");
import { IClient, GroupsAndEmails } from "../../../../../src/server/lib/ldap/IClient";
import Sinon = require("sinon");

export class ClientStub implements IClient {
  openStub: Sinon.SinonStub;
  closeStub: Sinon.SinonStub;
  searchUserDnStub: Sinon.SinonStub;
  searchEmailsStub: Sinon.SinonStub;
  searchEmailsAndGroupsStub: Sinon.SinonStub;
  modifyPasswordStub: Sinon.SinonStub;

  constructor() {
    this.openStub = Sinon.stub();
    this.closeStub = Sinon.stub();
    this.searchUserDnStub = Sinon.stub();
    this.searchEmailsStub = Sinon.stub();
    this.searchEmailsAndGroupsStub = Sinon.stub();
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

  searchEmailsAndGroups(username: string): BluebirdPromise<GroupsAndEmails> {
    return this.searchEmailsAndGroupsStub(username);
  }

  modifyPassword(username: string, newPassword: string): BluebirdPromise<void> {
    return this.modifyPasswordStub(username, newPassword);
  }
}