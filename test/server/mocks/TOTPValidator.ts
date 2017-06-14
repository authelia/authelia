
import sinon = require("sinon");

export interface TOTPValidatorMock {
    validate: sinon.SinonStub;
}

export function TOTPValidatorMock(): TOTPValidatorMock {
    return {
        validate: sinon.stub()
    };
}
