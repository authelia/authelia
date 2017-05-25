
import sinon = require("sinon");

export interface LdapjsMock {
    createClient: sinon.SinonStub;
}

export interface LdapjsClientMock {
    bind: sinon.SinonStub;
    search: sinon.SinonStub;
    modify: sinon.SinonStub;
    on: sinon.SinonStub;
}

export function LdapjsMock(): LdapjsMock {
    return {
        createClient: sinon.stub()
    };
}

export function LdapjsClientMock(): LdapjsClientMock {
    return {
        bind: sinon.stub(),
        search: sinon.stub(),
        modify: sinon.stub(),
        on: sinon.stub()
    };
}