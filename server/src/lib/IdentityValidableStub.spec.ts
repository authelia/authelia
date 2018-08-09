
import Sinon = require("sinon");
import { IdentityValidable } from "./IdentityValidable";
import express = require("express");
import Bluebird = require("bluebird");
import { Identity } from "../../types/Identity";


export class IdentityValidableStub implements IdentityValidable {
    challengeStub: Sinon.SinonStub;
    preValidationInitStub: Sinon.SinonStub;
    postValidationInitStub: Sinon.SinonStub;
    preValidationResponseStub: Sinon.SinonStub;
    postValidationResponseStub: Sinon.SinonStub;
    mailSubjectStub: Sinon.SinonStub;

    constructor() {
        this.challengeStub = Sinon.stub();

        this.preValidationInitStub = Sinon.stub();
        this.postValidationResponseStub = Sinon.stub();

        this.preValidationResponseStub = Sinon.stub();
        this.postValidationResponseStub = Sinon.stub();

        this.mailSubjectStub = Sinon.stub();
    }

    challenge(): string {
        return this.challengeStub();
    }

    preValidationInit(req: Express.Request): Bluebird<Identity> {
        return this.preValidationInitStub(req);
    }

    postValidationInit(req: Express.Request): Bluebird<void> {
        return this.postValidationInitStub(req);
    }

    preValidationResponse(req: Express.Request, res: Express.Response): void {
        return this.preValidationResponseStub(req, res);
    }

    postValidationResponse(req: Express.Request, res: Express.Response): void {
        return this.postValidationResponseStub(req, res);
    }

    mailSubject(): string {
        return this.mailSubjectStub();
    }
}