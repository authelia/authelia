import Sinon = require("sinon");
import express = require("express");
import winston = require("winston");
import { UserDataStoreStub } from "./storage/UserDataStoreStub";
import { VARIABLES_KEY } from "../../../../src/server/lib/ServerVariablesHandler";

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
    accessController: Sinon.stub(),
    config: Sinon.stub(),
    ldapAuthenticator: Sinon.stub() as any,
    ldapEmailsRetriever: Sinon.stub() as any,
    ldapPasswordUpdater: Sinon.stub() as any,
    logger: winston,
    notifier: Sinon.stub(),
    regulator: Sinon.stub(),
    totpGenerator: Sinon.stub(),
    totpValidator: Sinon.stub(),
    u2f: Sinon.stub(),
    userDataStore: new UserDataStoreStub()
  };
  app.get = Sinon.stub().withArgs(VARIABLES_KEY).returns(mocks);
  return mocks;
}