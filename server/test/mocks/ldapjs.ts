
import Sinon = require("sinon");

export class LdapjsMock {
    createClientStub: sinon.SinonStub;

    constructor() {
        this.createClientStub = Sinon.stub();
    }

    createClient(params: any) {
        return this.createClientStub(params);
    }
}

export class LdapjsClientMock {
    bindStub: sinon.SinonStub;
    unbindStub: sinon.SinonStub;
    searchStub: sinon.SinonStub;
    modifyStub: sinon.SinonStub;
    onStub: sinon.SinonStub;

    constructor() {
        this.bindStub = Sinon.stub();
        this.unbindStub = Sinon.stub();
        this.searchStub = Sinon.stub();
        this.modifyStub = Sinon.stub();
        this.onStub = Sinon.stub();
    }

    bind() {
        return this.bindStub();
    }

    unbind() {
        return this.unbindStub();
    }

    search() {
        return this.searchStub();
    }

    modify() {
        return this.modifyStub();
    }

    on() {
        return this.onStub();
    }
}