
import Express = require("express");
import { UserDataStore } from "./storage/UserDataStore";
import { Winston } from "../../types/Dependencies";

import FirstFactorGet = require("./routes/firstfactor/get");
import SecondFactorGet = require("./routes/secondfactor/get");

import FirstFactorPost = require("./routes/firstfactor/post");
import LogoutGet = require("./routes/logout/get");
import VerifyGet = require("./routes/verify/get");
import TOTPSignGet = require("./routes/secondfactor/totp/sign/post");

import IdentityCheckMiddleware = require("./IdentityCheckMiddleware");

import TOTPRegistrationIdentityHandler from "./routes/secondfactor/totp/identity/RegistrationHandler";
import U2FRegistrationIdentityHandler from "./routes/secondfactor/u2f/identity/RegistrationHandler";
import ResetPasswordIdentityHandler from "./routes/password-reset/identity/PasswordResetHandler";

import U2FSignPost = require("./routes/secondfactor/u2f/sign/post");
import U2FSignRequestGet = require("./routes/secondfactor/u2f/sign_request/get");

import U2FRegisterPost = require("./routes/secondfactor/u2f/register/post");
import U2FRegisterRequestGet = require("./routes/secondfactor/u2f/register_request/get");

import ResetPasswordFormPost = require("./routes/password-reset/form/post");
import ResetPasswordRequestPost = require("./routes/password-reset/request/get");

import Error401Get = require("./routes/error/401/get");
import Error403Get = require("./routes/error/403/get");
import Error404Get = require("./routes/error/404/get");

import LoggedIn = require("./routes/loggedin/get");

import { ServerVariables } from "./ServerVariables";
import { IRequestLogger } from "./logging/IRequestLogger";

import Endpoints = require("../../../shared/api");

function withHeadersLogged(fn: (req: Express.Request, res: Express.Response) => void,
  logger: IRequestLogger) {
  return function (req: Express.Request, res: Express.Response) {
    logger.debug(req, "Headers = %s", JSON.stringify(req.headers));
    fn(req, res);
  };
}

export class RestApi {
  static setup(app: Express.Application, vars: ServerVariables): void {
    app.get(Endpoints.FIRST_FACTOR_GET, withHeadersLogged(FirstFactorGet.default(vars), vars.logger));
    app.get(Endpoints.SECOND_FACTOR_GET, withHeadersLogged(SecondFactorGet.default(vars), vars.logger));
    app.get(Endpoints.LOGOUT_GET, withHeadersLogged(LogoutGet.default, vars.logger));

    IdentityCheckMiddleware.register(app, Endpoints.SECOND_FACTOR_TOTP_IDENTITY_START_GET,
      Endpoints.SECOND_FACTOR_TOTP_IDENTITY_FINISH_GET,
      new TOTPRegistrationIdentityHandler(vars.logger, vars.userDataStore, vars.totpHandler), vars);

    IdentityCheckMiddleware.register(app, Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_GET,
      Endpoints.SECOND_FACTOR_U2F_IDENTITY_FINISH_GET,
      new U2FRegistrationIdentityHandler(vars.logger), vars);

    IdentityCheckMiddleware.register(app, Endpoints.RESET_PASSWORD_IDENTITY_START_GET,
      Endpoints.RESET_PASSWORD_IDENTITY_FINISH_GET,
      new ResetPasswordIdentityHandler(vars.logger, vars.ldapEmailsRetriever), vars);

    app.get(Endpoints.RESET_PASSWORD_REQUEST_GET, withHeadersLogged(ResetPasswordRequestPost.default, vars.logger));
    app.post(Endpoints.RESET_PASSWORD_FORM_POST, withHeadersLogged(ResetPasswordFormPost.default(vars), vars.logger));

    app.get(Endpoints.VERIFY_GET, withHeadersLogged(VerifyGet.default(vars), vars.logger));
    app.post(Endpoints.FIRST_FACTOR_POST, withHeadersLogged(FirstFactorPost.default(vars), vars.logger));
    app.post(Endpoints.SECOND_FACTOR_TOTP_POST, withHeadersLogged(TOTPSignGet.default(vars), vars.logger));

    app.get(Endpoints.SECOND_FACTOR_U2F_SIGN_REQUEST_GET, withHeadersLogged(U2FSignRequestGet.default(vars), vars.logger));
    app.post(Endpoints.SECOND_FACTOR_U2F_SIGN_POST, withHeadersLogged(U2FSignPost.default(vars), vars.logger));

    app.get(Endpoints.SECOND_FACTOR_U2F_REGISTER_REQUEST_GET, withHeadersLogged(U2FRegisterRequestGet.default(vars), vars.logger));
    app.post(Endpoints.SECOND_FACTOR_U2F_REGISTER_POST, withHeadersLogged(U2FRegisterPost.default(vars), vars.logger));

    app.get(Endpoints.ERROR_401_GET, withHeadersLogged(Error401Get.default, vars.logger));
    app.get(Endpoints.ERROR_403_GET, withHeadersLogged(Error403Get.default, vars.logger));
    app.get(Endpoints.ERROR_404_GET, withHeadersLogged(Error404Get.default, vars.logger));
    app.get(Endpoints.LOGGED_IN, withHeadersLogged(LoggedIn.default(vars), vars.logger));
  }
}
