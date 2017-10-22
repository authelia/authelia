
import express = require("express");
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";

export default function(req: express.Request, res: express.Response) {
  const redirect_param = req.query.redirect;
  const redirect_url = redirect_param || "/";
  AuthenticationSessionHandler.reset(req);
  res.redirect(redirect_url);
}