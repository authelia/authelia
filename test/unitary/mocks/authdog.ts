
import sinon = require("sinon");
import authdog = require("authdog");

export interface AuthdogMock {
    startRegistration: sinon.SinonStub;
    finishRegistration: sinon.SinonStub;
    startAuthentication: sinon.SinonStub;
    finishAuthentication: sinon.SinonStub;
}

export function AuthdogMock(): AuthdogMock {
    return {
        startRegistration: sinon.stub(),
        finishAuthentication: sinon.stub(),
        startAuthentication: sinon.stub(),
        finishRegistration: sinon.stub()
    };
}
