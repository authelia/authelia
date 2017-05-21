
import FirstFactor = require("./routes/FirstFactor");
import second_factor = require("./routes/second_factor");
import reset_password = require("./routes/reset_password");
import AuthenticationValidator = require("./routes/AuthenticationValidator");
import u2f_register_handler = require("./routes/u2f_register_handler");
import totp_register = require("./routes/totp_register");
import objectPath = require("object-path");

import express = require("express");

export = {
  login: serveLogin,
  logout: serveLogout,
  verify: AuthenticationValidator,
  first_factor: FirstFactor,
  second_factor: second_factor,
  reset_password: reset_password,
  u2f_register: u2f_register_handler,
  totp_register: totp_register,
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

