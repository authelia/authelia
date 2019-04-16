import * as Express from "express";
import { ServerVariables } from "../../../ServerVariables";
import { AuthenticationSessionHandler } from "../../../AuthenticationSessionHandler";
import * as ErrorReplies from "../../../ErrorReplies";
import * as UserMessage from "../../../UserMessages";


export default function(vars: ServerVariables) {
  return async function(req: Express.Request, res: Express.Response) {
    try {
      if (!(req.body && req.body.method)) {
        throw new Error("No 'method' key in request body");
      }

      const authSession = AuthenticationSessionHandler.get(req, vars.logger);
      await vars.userDataStore.savePrefered2FAMethod(authSession.userid, req.body.method);
      res.status(204);
      res.send();
    } catch (err) {
      ErrorReplies.replyWithError200(req, res, vars.logger, UserMessage.OPERATION_FAILED)(err);
    }
  };
}