import Express = require("express");
import BluebirdPromise = require("bluebird");
import ObjectPath = require("object-path");
import { ServerVariables } from "../../ServerVariables";
import { AuthenticationSession }
  from "../../../../types/AuthenticationSession";
import { DomainExtractor } from "../../utils/DomainExtractor";
import { MethodCalculator } from "../../authentication/MethodCalculator";
import AccessControl from "./access_control";

export default function (req: Express.Request, res: Express.Response,
  vars: ServerVariables, authorizationHeader: string)
  : BluebirdPromise<{ username: string, groups: string[] }> {
  let username: string;
  let groups: string[];
  let domain: string;
  let originalUri: string;

  return new BluebirdPromise<[string, string]>(function (resolve, reject) {
    const originalUrl = ObjectPath.get<Express.Request, string>(req, "headers.x-original-url");
    domain = DomainExtractor.fromUrl(originalUrl);
    originalUri =
      ObjectPath.get<Express.Request, string>(req, "headers.x-original-uri");
    const authenticationMethod =
      MethodCalculator.compute(vars.config.authentication_methods, domain);

    if (authenticationMethod != "single_factor") {
      reject(new Error("This domain is not protected with single factor. " +
        "You cannot log in with basic authentication."));
      return;
    }

    const base64Re = new RegExp("^Basic ((?:[A-Za-z0-9+/]{4})*" +
      "(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?)$");
    const isTokenValidBase64 = base64Re.test(authorizationHeader);

    if (!isTokenValidBase64) {
      reject(new Error("No valid base64 token found in the header"));
      return;
    }

    const tokenMatches = authorizationHeader.match(base64Re);
    const base64Token = tokenMatches[1];
    const decodedToken = Buffer.from(base64Token, "base64").toString();
    const splittedToken = decodedToken.split(":");

    if (splittedToken.length != 2) {
      reject(new Error(
        "The authorization token is invalid. Expecting 'userid:password'"));
      return;
    }

    username = splittedToken[0];
    const password = splittedToken[1];
    resolve([username, password]);
  })
    .then(function ([userid, password]) {
      return vars.ldapAuthenticator.authenticate(userid, password);
    })
    .then(function (groupsAndEmails) {
      groups = groupsAndEmails.groups;
      return AccessControl(req, vars, domain, originalUri, username, groups);
    })
    .then(function () {
      return BluebirdPromise.resolve({
        username: username,
        groups: groups
      });
    })
    .catch(function (err: Error) {
      return BluebirdPromise.reject(
        new Error("Unable to authenticate the user with basic auth. Cause: "
          + err.message));
    });
}