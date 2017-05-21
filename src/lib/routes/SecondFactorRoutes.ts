
import DenyNotLogged = require("./DenyNotLogged");
import U2FRoutes = require("./U2FRoutes");
import TOTPAuthenticator = require("./TOTPAuthenticator");

import express = require("express");

interface SecondFactorRoutes {
  totp: express.RequestHandler;
  u2f: {
    register_request: express.RequestHandler;
    register: express.RequestHandler;
    sign_request: express.RequestHandler;
    sign: express.RequestHandler;
  };
}

export = {
  totp: DenyNotLogged(TOTPAuthenticator),
  u2f: {
    register_request: U2FRoutes.register_request,
    register: U2FRoutes.register,

    sign_request: DenyNotLogged(U2FRoutes.sign_request),
    sign: DenyNotLogged(U2FRoutes.sign),
  }
} as SecondFactorRoutes;

