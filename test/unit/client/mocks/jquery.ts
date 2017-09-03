
import sinon = require("sinon");
import jquery = require("jquery");


export interface JQueryMock extends sinon.SinonStub {
    get: sinon.SinonStub;
    post: sinon.SinonStub;
    ajax: sinon.SinonStub;
    notify: sinon.SinonStub;
}

export interface JQueryElementsMock {
    ready: sinon.SinonStub;
    show: sinon.SinonStub;
    hide: sinon.SinonStub;
    html: sinon.SinonStub;
    addClass: sinon.SinonStub;
    removeClass: sinon.SinonStub;
    fadeIn: sinon.SinonStub;
    fadeOut: sinon.SinonStub;
    on: sinon.SinonStub;
}

export interface JQueryDeferredMock {
    done: sinon.SinonStub;
    fail: sinon.SinonStub;
}

export function JQueryMock(): { jquery: JQueryMock, element: JQueryElementsMock } {
    const jquery = sinon.stub() as any;
    const jqueryInstance: JQueryElementsMock = {
        ready: sinon.stub(),
        show: sinon.stub(),
        hide: sinon.stub(),
        html: sinon.stub(),
        addClass: sinon.stub(),
        removeClass: sinon.stub(),
        fadeIn: sinon.stub(),
        fadeOut: sinon.stub(),
        on: sinon.stub()
    };
    jquery.ajax = sinon.stub();
    jquery.get = sinon.stub();
    jquery.post = sinon.stub();
    jquery.notify = sinon.stub();
    jquery.returns(jqueryInstance);
    return {
        jquery: jquery,
        element: jqueryInstance
    };
}

export function JQueryDeferredMock(): JQueryDeferredMock {
    return {
        done: sinon.stub(),
        fail: sinon.stub()
    };
}
