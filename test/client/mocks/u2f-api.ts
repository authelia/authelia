
import sinon = require("sinon");

export interface U2FApiMock {
    sign: sinon.SinonStub;
    register: sinon.SinonStub;
}

export function U2FApiMock(): U2FApiMock {
    return {
        sign: sinon.stub(),
        register: sinon.stub()
    };
}