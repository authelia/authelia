
import express = require("express");
import AuthenticationSession = require("../../AuthenticationSession");

export default function(req: express.Request, res: express.Response) {
  const redirect_param = req.query.redirect;
  const redirect_url = redirect_param || "/";
  AuthenticationSession.reset(req);
  res.redirect(redirect_url);
}