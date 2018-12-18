import Express = require("express");

import FirstFactorGet = require("../routes/firstfactor/get");
import SecondFactorGet = require("../routes/secondfactor/get");

import FirstFactorPost = require("../routes/firstfactor/post");
import LogoutGet = require("../routes/logout/get");
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
import ResetPasswordRequestPost = require("../routes/password-reset/request/get");

import Error401Get = require("../routes/error/401/get");
import Error403Get = require("../routes/error/403/get");
import Error404Get = require("../routes/error/404/get");

import LoggedIn = require("../routes/loggedin/get");

import { ServerVariables } from "../ServerVariables";
import Endpoints = require("../../../../shared/api");
import { RequireValidatedFirstFactor } from "./middlewares/RequireValidatedFirstFactor";

function setupTotp(app: Express.Application, vars: ServerVariables) {
  app.post(Endpoints.SECOND_FACTOR_TOTP_POST,
    RequireValidatedFirstFactor.middleware(vars.logger),
    TOTPSignGet.default(vars));

  app.get(Endpoints.SECOND_FACTOR_TOTP_IDENTITY_START_GET,
    RequireValidatedFirstFactor.middleware(vars.logger));

  app.get(Endpoints.SECOND_FACTOR_TOTP_IDENTITY_FINISH_GET,
    RequireValidatedFirstFactor.middleware(vars.logger));

  IdentityCheckMiddleware.register(app,
    Endpoints.SECOND_FACTOR_TOTP_IDENTITY_START_GET,
    Endpoints.SECOND_FACTOR_TOTP_IDENTITY_FINISH_GET,
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

  app.get(Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_GET,
    RequireValidatedFirstFactor.middleware(vars.logger));

  app.get(Endpoints.SECOND_FACTOR_U2F_IDENTITY_FINISH_GET,
    RequireValidatedFirstFactor.middleware(vars.logger));

  IdentityCheckMiddleware.register(app,
    Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_GET,
    Endpoints.SECOND_FACTOR_U2F_IDENTITY_FINISH_GET,
    new U2FRegistrationIdentityHandler(vars.logger), vars);
}

function setupResetPassword(app: Express.Application, vars: ServerVariables) {
  IdentityCheckMiddleware.register(app,
    Endpoints.RESET_PASSWORD_IDENTITY_START_GET,
    Endpoints.RESET_PASSWORD_IDENTITY_FINISH_GET,
    new ResetPasswordIdentityHandler(vars.logger, vars.usersDatabase),
    vars);

  app.get(Endpoints.RESET_PASSWORD_REQUEST_GET,
    ResetPasswordRequestPost.default);
  app.post(Endpoints.RESET_PASSWORD_FORM_POST,
    ResetPasswordFormPost.default(vars));
}

function setupErrors(app: Express.Application, vars: ServerVariables) {
  app.get(Endpoints.ERROR_401_GET, Error401Get.default(vars));
  app.get(Endpoints.ERROR_403_GET, Error403Get.default(vars));
  app.get(Endpoints.ERROR_404_GET, Error404Get.default);
}

export class RestApi {
  static setup(app: Express.Application, vars: ServerVariables): void {
    app.get(Endpoints.FIRST_FACTOR_GET, FirstFactorGet.default(vars));

    app.get(Endpoints.SECOND_FACTOR_GET,
      RequireValidatedFirstFactor.middleware(vars.logger),
      SecondFactorGet.default(vars));

    app.get(Endpoints.LOGOUT_GET, LogoutGet.default(vars));

    app.get(Endpoints.VERIFY_GET, VerifyGet.default(vars));
    app.post(Endpoints.FIRST_FACTOR_POST, FirstFactorPost.default(vars));

    setupTotp(app, vars);
    setupU2f(app, vars);
    setupResetPassword(app, vars);
    setupErrors(app, vars);

    app.get(Endpoints.LOGGED_IN,
      RequireValidatedFirstFactor.middleware(vars.logger),
      LoggedIn.default(vars));
  }
}
