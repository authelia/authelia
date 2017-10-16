import BluebirdPromise = require("bluebird");
import { IClient } from "../../../src/lib/ldap/IClient";
import { IPasswordUpdater } from "../../../src/lib/ldap/IPasswordUpdater";
import Sinon = require("sinon");

export class PasswordUpdaterStub implements IPasswordUpdater {
  updatePasswordStub: Sinon.SinonStub;

  constructor() {
    this.updatePasswordStub = Sinon.stub();
  }

  updatePassword(username: string, newPassword: string): BluebirdPromise<void> {
    return this.updatePasswordStub(username, newPassword);
  }
}