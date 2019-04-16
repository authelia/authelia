import Express = require("express");
import { ServerVariables } from "../../ServerVariables";
import { URLDecomposer } from "../../utils/URLDecomposer";
import { Level } from "../../authentication/Level";
import GetHeader from "../../utils/GetHeader";
import { HEADER_PROXY_AUTHORIZATION } from "../../constants";
import setUserAndGroupsHeaders from "./SetUserAndGroupsHeaders";
import CheckAuthorizations from "./CheckAuthorizations";
import RequestUrlGetter from "../../utils/RequestUrlGetter";

export default async function(req: Express.Request, res: Express.Response,
  vars: ServerVariables)
  : Promise<void> {
  const authorizationValue = GetHeader(req, HEADER_PROXY_AUTHORIZATION);

  if (!authorizationValue.startsWith("Basic ")) {
    throw new Error("The authorization header should be of the form 'Basic XXXXXX'");
  }

  const base64Re = new RegExp("^Basic ((?:[A-Za-z0-9+/]{4})*" +
    "(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?)$");
  const isTokenValidBase64 = base64Re.test(authorizationValue);

  if (!isTokenValidBase64) {
    throw new Error("No valid base64 token found in the header");
  }

  const tokenMatches = authorizationValue.match(base64Re);
  const base64Token = tokenMatches[1];
  const decodedToken = Buffer.from(base64Token, "base64").toString();
  const splittedToken = decodedToken.split(":");

  if (splittedToken.length != 2) {
    throw new Error("The authorization token is invalid. Expecting 'userid:password'");
  }

  const username = splittedToken[0];
  const password = splittedToken[1];
  const groupsAndEmails = await vars.usersDatabase.checkUserPassword(username, password);

  const uri = RequestUrlGetter.getOriginalUrl(req);
  const urlDecomposition = URLDecomposer.fromUrl(uri);

  CheckAuthorizations(vars.authorizer, urlDecomposition.domain, urlDecomposition.path,
    username, groupsAndEmails.groups, req.ip, Level.ONE_FACTOR);
  setUserAndGroupsHeaders(res, username, groupsAndEmails.groups);
}
