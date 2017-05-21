
import sinon = require("sinon");
import { IdentityValidable } from "../../../src/lib/IdentityValidator";
import express = require("express");
import BluebirdPromise = require("bluebird");
import { Identity } from "../../../src/types/Identity";


export interface IdentityValidableMock {
    challenge: sinon.SinonStub;
    templateName: sinon.SinonStub;
    preValidation: sinon.SinonStub;
    mailSubject: sinon.SinonStub;
}

export function IdentityValidableMock() {
    return {
        challenge: sinon.stub(),
        templateName: sinon.stub(),
        preValidation: sinon.stub(),
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