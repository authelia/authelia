import Sinon = require("sinon");
import { IAccessController } from "./IAccessController";

export class AccessControllerStub implements IAccessController {
  isAccessAllowedMock: Sinon.SinonStub;

  constructor() {
    this.isAccessAllowedMock = Sinon.stub();
  }

  isAccessAllowed(domain: string, resource: string, user: string, groups: string[]): boolean {
    return this.isAccessAllowedMock(domain, resource, user, groups);
  }
}
