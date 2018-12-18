
import sinon = require("sinon");

export interface U2FMock {
    request: sinon.SinonStub;
    checkSignature: sinon.SinonStub;
    checkRegistration: sinon.SinonStub;
}

export function U2FMock(): U2FMock {
    return {
        request: sinon.stub(),
        checkSignature: sinon.stub(),
        checkRegistration: sinon.stub()
    };
}
