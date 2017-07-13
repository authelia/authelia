
import sinon = require("sinon");

export interface LdapClientMock {
    checkPassword: sinon.SinonStub;
    retrieveEmails: sinon.SinonStub;
    retrieveGroups: sinon.SinonStub;
    search: sinon.SinonStub;
    updatePassword: sinon.SinonStub;
}

export function LdapClientMock(): LdapClientMock {
    return {
        checkPassword: sinon.stub(),
        retrieveEmails: sinon.stub(),
        retrieveGroups: sinon.stub(),
        search: sinon.stub(),
        updatePassword: sinon.stub()
    };
}
