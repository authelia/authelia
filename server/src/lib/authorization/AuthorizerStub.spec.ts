import Sinon = require("sinon");
import { IAuthorizer } from "./IAuthorizer";
import { Level } from "./Level";

export class AuthorizerStub implements IAuthorizer {
  authorizationMock: Sinon.SinonStub;

  constructor() {
    this.authorizationMock = Sinon.stub();
  }

  authorization(domain: string, resource: string, user: string, groups: string[]): Level {
    return this.authorizationMock(domain, resource, user, groups);
  }
}
