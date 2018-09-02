import { IUsersDatabase } from "../backends/IUsersDatabase";
import Bluebird = require("bluebird");
import express = require("express");
import { ServerVariables } from "../../ServerVariables";

export interface IWhitelistHandler {
  isWhitelisted(ip: string, usersDatabase: IUsersDatabase): Bluebird<string>;
  loginWhitelistUser(user: string, req: express.Request, vars: ServerVariables): Bluebird<void>;
}