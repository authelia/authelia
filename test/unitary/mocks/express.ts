
import sinon = require("sinon");

export interface RequestMock {
    app?: any;
    body?: any;
    session?: any;
    headers?: any;
}

export interface ResponseMock {
    send: sinon.SinonStub | sinon.SinonSpy;
    status: sinon.SinonStub;
    json: sinon.SinonStub;
}

export function RequestMock(): RequestMock {
    return {
        app: {
            get: sinon.stub()
        }
    };
}
export function ResponseMock(): ResponseMock {
    return {
        send: sinon.stub(),
        status: sinon.stub(),
        json: sinon.stub()
    };
}
