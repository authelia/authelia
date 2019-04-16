import * as Express from "express";
import { ServerVariables } from "../../../ServerVariables";
import { AuthenticationSessionHandler } from "../../../AuthenticationSessionHandler";
import * as ErrorReplies from "../../../ErrorReplies";
import * as UserMessage from "../../../UserMessages";
import redirect from "../redirect";
import { Level } from "../../../authentication/Level";
import { DuoPushConfiguration } from "../../../configuration/schema/DuoPushConfiguration";
import GetHeader from "../../../utils/GetHeader";
import { HEADER_X_TARGET_URL } from "../../../constants";
const DuoApi = require("@duosecurity/duo_api");

interface DuoResponse {
  response: {
    result: "allow" | "deny";
    status: "allow" | "deny" | "fraud";
    status_msg: string;
  };
  stat: "OK" | "FAIL";
}

function triggerAuth(username: string, config: DuoPushConfiguration, req: Express.Request): Promise<DuoResponse> {
  return new Promise((resolve, reject) => {
    const clientIP = req.ip;
    const targetURL = GetHeader(req, HEADER_X_TARGET_URL);
    const client = new DuoApi.Client(config.integration_key, config.secret_key, config.hostname);
    const timer = setTimeout(() => reject(new Error("Call to duo push API timed out.")), 60000);
    client.jsonApiCall("POST", "/auth/v2/auth", { username, ipaddr: clientIP, factor: "push", device: "auto", pushinfo: `target%20url=${targetURL}`}, (data: DuoResponse) => {
      clearTimeout(timer);
      resolve(data);
    });
  });
}


export default function(vars: ServerVariables) {
  return async function(req: Express.Request, res: Express.Response) {
    try {
      if (!vars.config.duo_api) {
        throw new Error("Duo Push Notification is not configured.");
      }

      const authSession = AuthenticationSessionHandler.get(req, vars.logger);
      const authRes = await triggerAuth(authSession.userid, vars.config.duo_api, req);
      if (authRes.response.result !== "allow") {
        throw new Error("User denied access.");
      }
      vars.logger.debug(req, "Access allowed by user via Duo Push.");
      authSession.authentication_level = Level.TWO_FACTOR;
      await redirect(vars)(req, res);
    } catch (err) {
      ErrorReplies.replyWithError200(req, res, vars.logger, UserMessage.OPERATION_FAILED)(err);
    }
  };
}