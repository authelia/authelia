import Bluebird = require("bluebird");
import Sinon = require("sinon");
import { IRegulator } from "./IRegulator";


export class RegulatorStub implements IRegulator {
    markStub: Sinon.SinonStub;
    regulateStub: Sinon.SinonStub;

    constructor() {
        this.markStub = Sinon.stub();
        this.regulateStub = Sinon.stub();
    }

    mark(userId: string, isAuthenticationSuccessful: boolean): Bluebird<void>  {
        return this.markStub(userId, isAuthenticationSuccessful);
    }

    regulate(userId: string): Bluebird<void> {
        return this.regulateStub(userId);
    }
}
