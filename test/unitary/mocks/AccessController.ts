
import sinon = require("sinon");

export interface AccessControllerMock {
    isDomainAllowedForUser: sinon.SinonStub;
}

export function AccessControllerMock() {
    return {
        isDomainAllowedForUser: sinon.stub()
    };
}
