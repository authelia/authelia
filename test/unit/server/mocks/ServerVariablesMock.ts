import sinon = require("sinon");
import express = require("express");
import winston = require("winston");
import { UserDataStoreStub } from "./storage/UserDataStoreStub";
import { ServerVariables, VARIABLES_KEY }  from "../../../../src/server/lib/ServerVariablesHandler";

export interface ServerVariablesMock {
    logger: any;
    ldapAuthenticator: any;
    ldapEmailsRetriever: any;
    ldapPasswordUpdater: any;
    totpValidator: any;
    totpGenerator: any;
    u2f: any;
    userDataStore: UserDataStoreStub;
    notifier: any;
    regulator: any;
    config: any;
    accessController: any;
}


export function mock(app: express.Application): ServerVariablesMock {
    const mocks: ServerVariablesMock = {
        accessController: sinon.stub(),
        config: sinon.stub(),
        ldapAuthenticator: sinon.stub() as any,
        ldapEmailsRetriever: sinon.stub() as any,
        ldapPasswordUpdater: sinon.stub() as any,
        logger: winston,
        notifier: sinon.stub(),
        regulator: sinon.stub(),
        totpGenerator: sinon.stub(),
        totpValidator: sinon.stub(),
        u2f: sinon.stub(),
        userDataStore: new UserDataStoreStub()
    };
    app.get = sinon.stub().withArgs(VARIABLES_KEY).returns(mocks);
    return mocks;
}