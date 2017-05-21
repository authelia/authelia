
import U2FRegistrationProcess = require("./U2FRegistrationProcess");
import U2FAuthenticationProcess = require("./U2FAuthenticationProcess");

import express = require("express");

interface U2FRoutes {
  register_request: express.RequestHandler;
  register: express.RequestHandler;
  sign_request: express.RequestHandler;
  sign: express.RequestHandler;
}

export = {
  register_request: U2FRegistrationProcess.register_request,
  register: U2FRegistrationProcess.register,
  sign_request: U2FAuthenticationProcess.sign_request,
  sign: U2FAuthenticationProcess.sign,
} as U2FRoutes;
