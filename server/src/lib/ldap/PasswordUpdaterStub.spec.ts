import BluebirdPromise = require("bluebird");
import { IClient } from "./IClient";
import { IPasswordUpdater } from "./IPasswordUpdater";
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