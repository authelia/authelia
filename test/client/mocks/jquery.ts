
import sinon = require("sinon");
import jquery = require("jquery");


export interface JQueryMock extends sinon.SinonStub {
    get: sinon.SinonStub;
    post: sinon.SinonStub;
    ajax: sinon.SinonStub;
    notify: sinon.SinonStub;
}

export interface JQueryDeferredMock {
    done: sinon.SinonStub;
    fail: sinon.SinonStub;
}

export function JQueryMock(): JQueryMock {
    const jquery = sinon.stub() as any;
    const jqueryInstance = {
        ready: sinon.stub(),
        show: sinon.stub(),
        hide: sinon.stub(),
        on: sinon.stub()
    };
    jquery.ajax = sinon.stub();
    jquery.get = sinon.stub();
    jquery.post = sinon.stub();
    jquery.notify = sinon.stub();
    jquery.returns(jqueryInstance);
    return jquery;
}

export function JQueryDeferredMock(): JQueryDeferredMock {
    return {
        done: sinon.stub(),
        fail: sinon.stub()
    };
}
