import Sinon = require("sinon");
import Bluebird = require("bluebird");
import express = require("express");
import { IWhitelistHandler } from "./IWhitelistHandler";
import { IUsersDatabase } from "../backends/IUsersDatabase";
import { ServerVariables } from "../../ServerVariables";

export class WhitelistHandlerStub implements IWhitelistHandler {
  isWhitelistedStub: Sinon.SinonStub;
  loginWhitelistUserStub: Sinon.SinonStub;

  constructor() {
    this.isWhitelistedStub = Sinon.stub();
    this.loginWhitelistUserStub = Sinon.stub();
  }

  isWhitelisted(ip: string, usersDatabase: IUsersDatabase): Bluebird<string> {
    return this.isWhitelistedStub(ip, usersDatabase);
  }

  loginWhitelistUser(user: string, req: express.Request, vars: ServerVariables): Bluebird<void> {
    return this.loginWhitelistUserStub(user, req, vars);
  }
}