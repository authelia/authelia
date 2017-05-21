
import sinon = require("sinon");


export interface AuthenticationRegulatorMock {
    mark: sinon.SinonStub;
    regulate: sinon.SinonStub;
}

export function AuthenticationRegulatorMock() {
    return {
        mark: sinon.stub(),
        regulate: sinon.stub()
    };
}
