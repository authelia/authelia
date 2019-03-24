import * as Express from "express";
import { ServerVariables } from "../../../ServerVariables";
import { AuthenticationSessionHandler } from "../../../AuthenticationSessionHandler";
import * as ErrorReplies from "../../../ErrorReplies";
import * as UserMessage from "../../../../../../shared/UserMessages";
import redirect from "../redirect";
import { Level } from "../../../authentication/Level";
import { DuoPushConfiguration } from "../../../configuration/schema/DuoPushConfiguration";
const DuoApi = require("@duosecurity/duo_api");

interface DuoResponse {
  response: {
    result: "allow" | "deny";
    status: "allow" | "deny" | "fraud";
    status_msg: string;
  };
  stat: "OK" | "FAIL";
}

function triggerAuth(username: string, config: DuoPushConfiguration): Promise<DuoResponse> {
  return new Promise((resolve, reject) => {
    const client = new DuoApi.Client(config.integration_key, config.secret_key, config.hostname);
    const timer = setTimeout(() => reject(new Error("Call to duo push API timed out.")), 60000);
    client.jsonApiCall("POST", "/auth/v2/auth", { username, factor: "push", device: "auto" }, (data: DuoResponse) => {
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
      const authRes = await triggerAuth(authSession.userid, vars.config.duo_api);
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