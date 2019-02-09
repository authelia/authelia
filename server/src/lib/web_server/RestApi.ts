import Express = require("express");

import FirstFactorPost = require("../routes/firstfactor/post");
import LogoutPost from "../routes/logout/post";
import StateGet from "../routes/state/get";
import RedirectPost from "../routes/redirect/post";
import VerifyGet = require("../routes/verify/get");
import TOTPSignGet = require("../routes/secondfactor/totp/sign/post");

import IdentityCheckMiddleware = require("../IdentityCheckMiddleware");

import TOTPRegistrationIdentityHandler from "../routes/secondfactor/totp/identity/RegistrationHandler";
import U2FRegistrationIdentityHandler from "../routes/secondfactor/u2f/identity/RegistrationHandler";
import ResetPasswordIdentityHandler from "../routes/password-reset/identity/PasswordResetHandler";

import U2FSignPost = require("../routes/secondfactor/u2f/sign/post");
import U2FSignRequestGet = require("../routes/secondfactor/u2f/sign_request/get");

import U2FRegisterPost = require("../routes/secondfactor/u2f/register/post");
import U2FRegisterRequestGet = require("../routes/secondfactor/u2f/register_request/get");

import ResetPasswordFormPost = require("../routes/password-reset/form/post");

import { ServerVariables } from "../ServerVariables";
import Endpoints = require("../../../../shared/api");
import { RequireValidatedFirstFactor } from "./middlewares/RequireValidatedFirstFactor";

function setupTotp(app: Express.Application, vars: ServerVariables) {
  app.post(Endpoints.SECOND_FACTOR_TOTP_POST,
    RequireValidatedFirstFactor.middleware(vars.logger),
    TOTPSignGet.default(vars));

  app.post(Endpoints.SECOND_FACTOR_TOTP_IDENTITY_START_POST,
    RequireValidatedFirstFactor.middleware(vars.logger));

  app.post(Endpoints.SECOND_FACTOR_TOTP_IDENTITY_FINISH_POST,
    RequireValidatedFirstFactor.middleware(vars.logger));

  IdentityCheckMiddleware.register(app,
    Endpoints.SECOND_FACTOR_TOTP_IDENTITY_START_POST,
    Endpoints.SECOND_FACTOR_TOTP_IDENTITY_FINISH_POST,
    new TOTPRegistrationIdentityHandler(vars.logger,
      vars.userDataStore, vars.totpHandler, vars.config.totp),
    vars);
}

function setupU2f(app: Express.Application, vars: ServerVariables) {
  app.get(Endpoints.SECOND_FACTOR_U2F_SIGN_REQUEST_GET,
    RequireValidatedFirstFactor.middleware(vars.logger),
    U2FSignRequestGet.default(vars));

  app.post(Endpoints.SECOND_FACTOR_U2F_SIGN_POST,
    RequireValidatedFirstFactor.middleware(vars.logger),
    U2FSignPost.default(vars));

  app.get(Endpoints.SECOND_FACTOR_U2F_REGISTER_REQUEST_GET,
    RequireValidatedFirstFactor.middleware(vars.logger),
    U2FRegisterRequestGet.default(vars));

  app.post(Endpoints.SECOND_FACTOR_U2F_REGISTER_POST,
    RequireValidatedFirstFactor.middleware(vars.logger),
    U2FRegisterPost.default(vars));

  app.post(Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_POST,
    RequireValidatedFirstFactor.middleware(vars.logger));

  app.post(Endpoints.SECOND_FACTOR_U2F_IDENTITY_FINISH_POST,
    RequireValidatedFirstFactor.middleware(vars.logger));

  IdentityCheckMiddleware.register(app,
    Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_POST,
    Endpoints.SECOND_FACTOR_U2F_IDENTITY_FINISH_POST,
    new U2FRegistrationIdentityHandler(vars.logger), vars);
}

function setupResetPassword(app: Express.Application, vars: ServerVariables) {
  IdentityCheckMiddleware.register(app,
    Endpoints.RESET_PASSWORD_IDENTITY_START_GET,
    Endpoints.RESET_PASSWORD_IDENTITY_FINISH_GET,
    new ResetPasswordIdentityHandler(vars.logger, vars.usersDatabase),
    vars);

  app.post(Endpoints.RESET_PASSWORD_FORM_POST,
    ResetPasswordFormPost.default(vars));
}

export class RestApi {
  static setup(app: Express.Application, vars: ServerVariables): void {
    app.get(Endpoints.STATE_GET, StateGet(vars));
    app.post(Endpoints.REDIRECT_POST, RedirectPost(vars));

    app.post(Endpoints.LOGOUT_POST, LogoutPost(vars));

    app.get(Endpoints.VERIFY_GET, VerifyGet.default(vars));
    app.post(Endpoints.FIRST_FACTOR_POST, FirstFactorPost.default(vars));

    setupTotp(app, vars);
    setupU2f(app, vars);
    setupResetPassword(app, vars);
  }
}
