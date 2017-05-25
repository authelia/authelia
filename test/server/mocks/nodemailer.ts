
import sinon = require("sinon");

export interface NodemailerMock {
    createTransport: sinon.SinonStub;
}

export function NodemailerMock(): NodemailerMock {
    return {
        createTransport: sinon.stub()
    };
}

export interface NodemailerTransporterMock {
    sendMail: sinon.SinonStub;
}

export function NodemailerTransporterMock() {
    return {
        sendMail: sinon.stub()
    };
}
