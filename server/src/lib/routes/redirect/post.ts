import * as Express from "express";
import { ServerVariables } from "../../ServerVariables";
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { Level } from "../../authentication/Level";
import * as URLParse from "url-parse";

export default function (vars: ServerVariables) {
  return function (req: Express.Request, res: Express.Response) {
    if (!req.body.url) {
      res.status(400);
      vars.logger.error(req, "Provide url for verification to be done.");
      return;
    }

    const authSession = AuthenticationSessionHandler.get(req, vars.logger);
    const url = new URLParse(req.body.url);

    const urlInDomain = url.hostname.endsWith(vars.config.session.domain);
    vars.logger.debug(req, "Check domain %s is in url %s.", vars.config.session.domain, url.hostname);
    const sufficientPermissions = authSession.authentication_level >= Level.TWO_FACTOR;

    vars.logger.debug(req, "Check that protocol %s is HTTPS.", url.protocol);
    const protocolIsHttps = url.protocol === "https:";

    if (sufficientPermissions && urlInDomain && protocolIsHttps) {
      res.send("OK");
      return;
    }
    res.send("NOK");
  };
}
