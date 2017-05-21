
import FirstFactor = require("./routes/FirstFactor");
import SecondFactorRoutes = require("./routes/SecondFactorRoutes");
import PasswordReset = require("./routes/PasswordReset");
import AuthenticationValidator = require("./routes/AuthenticationValidator");
import U2FRegistration = require("./routes/U2FRegistration");
import TOTPRegistration = require("./routes/TOTPRegistration");
import objectPath = require("object-path");

import express = require("express");

export = {
  login: serveLogin,
  logout: serveLogout,
  verify: AuthenticationValidator,
  first_factor: FirstFactor,
  second_factor: SecondFactorRoutes,
  reset_password: PasswordReset,
  u2f_register: U2FRegistration,
  totp_register: TOTPRegistration,
};

function serveLogin(req: express.Request, res: express.Response) {
  if (!(objectPath.has(req, "session.auth_session"))) {
    req.session.auth_session = {};
    req.session.auth_session.first_factor = false;
    req.session.auth_session.second_factor = false;
  }
  res.render("login");
}

function serveLogout(req: express.Request, res: express.Response) {
  const redirect_param = req.query.redirect;
  const redirect_url = redirect_param || "/";
  req.session.auth_session = {
    first_factor: false,
    second_factor: false
  };
  res.redirect(redirect_url);
}

