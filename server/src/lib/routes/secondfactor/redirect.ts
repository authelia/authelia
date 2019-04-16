
import * as Express from "express";
import * as URLParse from "url-parse";
import { ServerVariables } from "../../ServerVariables";
import IsRedirectionSafe from "../../../lib/utils/IsRedirectionSafe";
import GetHeader from "../../utils/GetHeader";
import { HEADER_X_TARGET_URL } from "../../constants";


export default function (vars: ServerVariables) {
  return async function (req: Express.Request, res: Express.Response): Promise<void> {
    let redirectUrl = GetHeader(req, HEADER_X_TARGET_URL);

    if (!redirectUrl && vars.config.default_redirection_url) {
      redirectUrl = vars.config.default_redirection_url;
    }

    if ((redirectUrl && !IsRedirectionSafe(vars, new URLParse(redirectUrl)))
        || !redirectUrl) {
      res.status(204);
      res.send();
      return;
    }

    res.json({redirect: redirectUrl});
  };
}
