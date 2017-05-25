import sinon = require("sinon");
import express = require("express");
import {  ServerVariables, VARIABLES_KEY }  from "../../../src/server/lib/ServerVariables";

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


export function mock(app: express.Application): ServerVariablesMock {
    const mocks: ServerVariablesMock = {
        accessController: sinon.stub(),
        config: sinon.stub(),
        ldap: sinon.stub(),
        logger: sinon.stub(),
        notifier: sinon.stub(),
        regulator: sinon.stub(),
        totpGenerator: sinon.stub(),
        totpValidator: sinon.stub(),
        u2f: sinon.stub(),
        userDataStore: sinon.stub()
    };
    app.get = sinon.stub().withArgs(VARIABLES_KEY).returns(mocks);
    return mocks;
}