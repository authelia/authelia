
import sinon = require("sinon");

export interface LdapClientMock {
    bind: sinon.SinonStub;
    get_emails: sinon.SinonStub;
    get_groups: sinon.SinonStub;
    search_in_ldap: sinon.SinonStub;
    update_password: sinon.SinonStub;
}

export function LdapClientMock(): LdapClientMock {
    return {
        bind: sinon.stub(),
        get_emails: sinon.stub(),
        get_groups: sinon.stub(),
        search_in_ldap: sinon.stub(),
        update_password: sinon.stub()
    };
}
