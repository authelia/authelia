import * as Express from "express";
import { ServerVariables } from "../../../ServerVariables";
import { AuthenticationSessionHandler } from "../../../AuthenticationSessionHandler";
import * as ErrorReplies from "../../../ErrorReplies";
import * as UserMessage from "../../../UserMessages";


export default function(vars: ServerVariables) {
  return async function(req: Express.Request, res: Express.Response) {
    try {
      const authSession = AuthenticationSessionHandler.get(req, vars.logger);
      const method = await vars.userDataStore.retrievePrefered2FAMethod(authSession.userid);
      res.json({method});
    } catch (err) {
      ErrorReplies.replyWithError200(req, res, vars.logger, UserMessage.OPERATION_FAILED)(err);
    }
  };
}