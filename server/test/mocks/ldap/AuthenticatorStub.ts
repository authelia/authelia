import BluebirdPromise = require("bluebird");
import { IAuthenticator } from "../../../src/lib/ldap/IAuthenticator";
import { GroupsAndEmails } from "../../../src/lib/ldap/IClient";
import Sinon = require("sinon");

export class AuthenticatorStub implements IAuthenticator {
  authenticateStub: Sinon.SinonStub;

  constructor() {
    this.authenticateStub = Sinon.stub();
  }

  authenticate(username: string, password: string): BluebirdPromise<GroupsAndEmails> {
    return this.authenticateStub(username, password);
  }
}