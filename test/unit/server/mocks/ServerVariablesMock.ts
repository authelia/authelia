import sinon = require("sinon");
import express = require("express");
import {  ServerVariables, VARIABLES_KEY }  from "../../../../src/server/lib/ServerVariables";

export interface ServerVariablesMock {
    logger: any;
    ldap: any;
    totpValidator: any;
    totpGenerator: any;
    u2f: any;
    userDataStore: any;
    notifier: any;
    regulator: any;
    config: any;
    accessController: any;
}


export function mock(app: express.Application): ServerVariables {
    const mocks: ServerVariables = {
        accessController: sinon.stub() as any,
        config: sinon.stub() as any,
        ldapAuthenticator: sinon.stub() as any,
        ldapEmailsRetriever: sinon.stub() as any,
        ldapPasswordUpdater: sinon.stub() as any,
        logger: sinon.stub() as any,
        notifier: sinon.stub() as any,
        regulator: sinon.stub() as any,
        totpGenerator: sinon.stub() as any,
        totpValidator: sinon.stub() as any,
        u2f: sinon.stub() as any,
        userDataStore: sinon.stub() as any,
    };
    app.get = sinon.stub().withArgs(VARIABLES_KEY).returns(mocks);
    return mocks;
}