import Sinon = require("sinon");
import express = require("express");
import { RequestLoggerStub } from "./RequestLoggerStub";
import { UserDataStoreStub } from "./storage/UserDataStoreStub";
import { AuthenticationMethodCalculator } from "../../src/lib/AuthenticationMethodCalculator";
import { VARIABLES_KEY } from "../../src/lib/ServerVariablesHandler";

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
  authenticationMethodsCalculator: any;
}


export function mock(app: express.Application): ServerVariablesMock {
  const mocks: ServerVariablesMock = {
    accessController: Sinon.stub(),
    config: Sinon.stub(),
    ldapAuthenticator: Sinon.stub() as any,
    ldapEmailsRetriever: Sinon.stub() as any,
    ldapPasswordUpdater: Sinon.stub() as any,
    logger: new RequestLoggerStub(),
    notifier: Sinon.stub(),
    regulator: Sinon.stub(),
    totpGenerator: Sinon.stub(),
    totpValidator: Sinon.stub(),
    u2f: Sinon.stub(),
    userDataStore: new UserDataStoreStub(),
    authenticationMethodsCalculator: new AuthenticationMethodCalculator({
      default_method: "two_factor",
      per_subdomain_methods: {}
    })
  };
  app.get = Sinon.stub().withArgs(VARIABLES_KEY).returns(mocks);
  return mocks;
}