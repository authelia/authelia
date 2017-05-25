
import sinon = require("sinon");
import { IdentityValidable } from "../../../src/server/lib/IdentityCheckMiddleware";
import express = require("express");
import BluebirdPromise = require("bluebird");
import { Identity } from "../../../src/types/Identity";


export interface IdentityValidableMock {
    challenge: sinon.SinonStub;
    preValidationInit: sinon.SinonStub;
    preValidationResponse: sinon.SinonStub | sinon.SinonSpy;
    postValidationInit: sinon.SinonStub;
    postValidationResponse: sinon.SinonStub | sinon.SinonSpy;
    mailSubject: sinon.SinonStub;
}

export function IdentityValidableMock() {
    return {
        challenge: sinon.stub(),
        preValidationInit: sinon.stub(),
        preValidationResponse: sinon.stub(),
        postValidationInit: sinon.stub(),
        postValidationResponse: sinon.stub(),
        mailSubject: sinon.stub()
    };
}

export interface IdentityValidatorMock {
    consume_token: sinon.SinonStub;
    issue_token: sinon.SinonStub;
}

export function IdentityValidatorMock() {
    return {
        consume_token: sinon.stub(),
        issue_token: sinon.stub()
    };
}