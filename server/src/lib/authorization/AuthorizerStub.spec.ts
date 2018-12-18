import Sinon = require("sinon");
import { IAuthorizer } from "./IAuthorizer";
import { Level } from "./Level";
import { Object } from "./Object";
import { Subject } from "./Subject";

export class AuthorizerStub implements IAuthorizer {
  authorizationMock: Sinon.SinonStub;

  constructor() {
    this.authorizationMock = Sinon.stub();
  }

  authorization(object: Object, subject: Subject): Level {
    return this.authorizationMock(object, subject);
  }
}
