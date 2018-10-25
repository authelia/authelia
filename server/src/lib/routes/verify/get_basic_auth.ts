import Express = require("express");
import BluebirdPromise = require("bluebird");
import ObjectPath = require("object-path");
import { ServerVariables } from "../../ServerVariables";
import { AuthenticationSession }
  from "../../../../types/AuthenticationSession";
import AccessControl from "./access_control";
import { URLDecomposer } from "../../utils/URLDecomposer";
import { Level } from "../../authentication/Level";

export default function (req: Express.Request, res: Express.Response,
  vars: ServerVariables, authorizationHeader: string)
  : BluebirdPromise<{ username: string, groups: string[] }> {
  let username: string;
  const uri = ObjectPath.get<Express.Request, string>(req, "headers.x-original-url");
  const urlDecomposition = URLDecomposer.fromUrl(uri);

  return BluebirdPromise.resolve()
    .then(() => {
      const base64Re = new RegExp("^Basic ((?:[A-Za-z0-9+/]{4})*" +
        "(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?)$");
      const isTokenValidBase64 = base64Re.test(authorizationHeader);

      if (!isTokenValidBase64) {
        return BluebirdPromise.reject(new Error("No valid base64 token found in the header"));
      }

      const tokenMatches = authorizationHeader.match(base64Re);
      const base64Token = tokenMatches[1];
      const decodedToken = Buffer.from(base64Token, "base64").toString();
      const splittedToken = decodedToken.split(":");

      if (splittedToken.length != 2) {
        return BluebirdPromise.reject(new Error(
          "The authorization token is invalid. Expecting 'userid:password'"));
      }

      username = splittedToken[0];
      const password = splittedToken[1];
      return vars.usersDatabase.checkUserPassword(username, password);
    })
    .then(function (groupsAndEmails) {
      return AccessControl(req, vars, urlDecomposition.domain, urlDecomposition.path,
        username, groupsAndEmails.groups, Level.ONE_FACTOR)
        .then(() => BluebirdPromise.resolve({
          username: username,
          groups: groupsAndEmails.groups
        }));
    })
    .catch(function (err: Error) {
      return BluebirdPromise.reject(
        new Error("Unable to authenticate the user with basic auth. Cause: "
          + err.message));
    });
}