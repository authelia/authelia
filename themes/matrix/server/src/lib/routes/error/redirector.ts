import Express = require("express");
import { ServerVariables } from "../../ServerVariables";

export default function (req: Express.Request, vars: ServerVariables): string {
  let redirectionUrl: string;

  if (req.headers && req.headers["referer"])
    redirectionUrl = "" + req.headers["referer"];
  else if (vars.config.default_redirection_url)
    redirectionUrl = vars.config.default_redirection_url;

  return redirectionUrl;
}