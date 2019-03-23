import Express = require("express");
import { ServerVariables } from "../../ServerVariables";
import { AuthenticationSession }
  from "../../../../types/AuthenticationSession";
import { URLDecomposer } from "../../utils/URLDecomposer";
import GetHeader from "../../utils/GetHeader";
import {
  HEADER_X_ORIGINAL_URL,
} from "../../../../../shared/constants";
import { Level as AuthorizationLevel } from "../../authorization/Level";
import setUserAndGroupsHeaders from "./SetUserAndGroupsHeaders";
import CheckAuthorizations from "./CheckAuthorizations";
import CheckInactivity from "./CheckInactivity";


export default async function (req: Express.Request, res: Express.Response,
  vars: ServerVariables, authSession: AuthenticationSession | undefined)
  : Promise<void> {
  if (!authSession) {
    throw new Error("No cookie detected.");
  }

  const originalUrl = GetHeader(req, HEADER_X_ORIGINAL_URL);

  if (!originalUrl) {
    throw new Error("Cannot detect the original URL from headers.");
  }

  const d = URLDecomposer.fromUrl(originalUrl);

  const username = authSession.userid;
  const groups = authSession.groups;

  vars.logger.debug(req, "domain=%s, path=%s, user=%s, groups=%s", d.domain,
    d.path, (username) ? username : "unknown", (groups instanceof Array && groups.length > 0) ? groups.join(",") : "unknown");
  const authorizationLevel = CheckAuthorizations(vars.authorizer, d.domain, d.path, username, groups,
    authSession.authentication_level);

  if (authorizationLevel > AuthorizationLevel.BYPASS) {
    CheckInactivity(req, authSession, vars.config, vars.logger);
    setUserAndGroupsHeaders(res, username, groups);
  }
}