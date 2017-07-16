
import sinon = require("sinon");

export interface NotifierMock {
    notify: sinon.SinonStub;
}

export function NotifierMock(): NotifierMock {
    return {
        notify: sinon.stub()
    };
}
