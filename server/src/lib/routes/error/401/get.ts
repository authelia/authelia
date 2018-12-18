
import BluebirdPromise = require("bluebird");
import express = require("express");
import redirector from "../redirector";
import { ServerVariables } from "../../../ServerVariables";

export default function (vars: ServerVariables) {
  return function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    const redirectionUrl = redirector(req, vars);
    res.render("errors/401", {
      redirection_url: redirectionUrl
    });
    return BluebirdPromise.resolve();
  };
}
