import Express = require("express");
import { ServerVariables } from "../../ServerVariables";
import { AuthenticationSession }
  from "../../../../types/AuthenticationSession";
import setUserAndGroupsHeaders from "./SetUserAndGroupsHeaders";
import CheckAuthorizations from "./CheckAuthorizations";
import CheckInactivity from "./CheckInactivity";
import RequestUrlGetter from "../../utils/RequestUrlGetter";
import * as URLParse from "url-parse";


export default async function (req: Express.Request, res: Express.Response,
  vars: ServerVariables, authSession: AuthenticationSession | undefined)
  : Promise<void> {
  if (!authSession) {
    throw new Error("No cookie detected.");
  }

  const originalUrl = RequestUrlGetter.getOriginalUrl(req);

  if (!originalUrl) {
    throw new Error("Cannot detect the original URL from headers.");
  }

  const url = new URLParse(originalUrl);
  const username = authSession.userid;
  const groups = authSession.groups;

  vars.logger.debug(req, "domain=%s, path=%s, user=%s, groups=%s, ip=%s", url.hostname,
    url.pathname, (username) ? username : "unknown",
    (groups instanceof Array && groups.length > 0) ? groups.join(",") : "unknown", req.ip);

  CheckAuthorizations(vars.authorizer, url.hostname, url.pathname, username, groups,
    req.ip, authSession.authentication_level);
  CheckInactivity(req, authSession, vars.config, vars.logger);
  setUserAndGroupsHeaders(res, username, groups);
}
