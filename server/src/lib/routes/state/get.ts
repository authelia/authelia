import * as Express from "express";
import * as Bluebird from "bluebird";
import { ServerVariables } from "../../ServerVariables";
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";

export default function (vars: ServerVariables) {
  return function (req: Express.Request, res: Express.Response): Bluebird<void> {
    return new Bluebird(function (resolve, reject) {
      const authSession = AuthenticationSessionHandler.get(req, vars.logger);
      res.json({
        username: authSession.userid,
        authentication_level: authSession.authentication_level,
        default_redirection_url: vars.config.default_redirection_url,
      });
      resolve();
    });
  };
}
