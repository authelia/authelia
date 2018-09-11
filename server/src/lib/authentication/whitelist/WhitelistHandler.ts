import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { IWhitelistHandler } from "./IWhitelistHandler";
import { IUsersDatabase } from "../backends/IUsersDatabase";
import { ServerVariables } from "../../ServerVariables";
import Constants = require("../../../../../shared/constants");
import Bluebird = require("bluebird");
import express = require("express");
import ipRangeCheck = require("ip-range-check");

export const enum WhitelistValue {
  NOT_WHITELISTED,
  WHITELISTED,
  WHITELISTED_AND_AUTHENTICATED_FIRSTFACTOR,
  WHITELISTED_AND_AUTHENTICATED_SECONDFACTOR
}

export class WhitelistHandler implements IWhitelistHandler {
  isWhitelisted(ip: string, usersDatabase: IUsersDatabase): Bluebird<string> {
    // Get Users & Network Addresses
    return usersDatabase.getUsersWithNetworkAddresses()
      .then((users) => {
          // Search through users for a matching ip
          const user = users.find((user) => ipRangeCheck(ip, user.network_addresses));
          return Bluebird.resolve(user.user);
        }
      );
  }

  loginWhitelistUser(user: string, req: express.Request, vars: ServerVariables): Bluebird<void> {
    let authSession: AuthenticationSession;
    authSession = AuthenticationSessionHandler.get(req, vars.logger);
    authSession.userid = user;
    authSession.whitelisted = WhitelistValue.WHITELISTED;

    return vars.usersDatabase.getEmails(user)
      .then((emails) => {
        if (emails.length > 0)
          authSession.email = emails[0];
        return vars.usersDatabase.getGroups(user);
      })
      .then((groups) => {
        authSession.groups = groups;
        return Bluebird.resolve();
      });
  }
}