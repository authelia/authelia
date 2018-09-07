import Sinon = require("sinon");
import { IAccessController } from "./IAccessController";
import { WhitelistValue } from "../authentication/whitelist/WhitelistHandler";

export class AccessControllerStub implements IAccessController {
  isAccessAllowedMock: Sinon.SinonStub;

  constructor() {
    this.isAccessAllowedMock = Sinon.stub();
  }

  isAccessAllowed(domain: string, resource: string, user: string, groups: string[], whitelisted: WhitelistValue, secondFactorAuth: boolean): boolean {
    return this.isAccessAllowedMock(domain, resource, user, groups, whitelisted, secondFactorAuth);
  }
}
