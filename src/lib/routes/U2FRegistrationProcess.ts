
import u2f_register_handler = require("./U2FRegistration");
import objectPath = require("object-path");
import u2f_common = require("./u2f_common");
import BluebirdPromise = require("bluebird");
import express = require("express");
import authdog = require("../../types/authdog");

function register_request(req: express.Request, res: express.Response) {
  const logger = req.app.get("logger");
  const challenge = objectPath.get(req, "session.auth_session.identity_check.challenge");
  if (challenge != "u2f-register") {
    res.status(403);
    res.send();
    return;
  }

  const u2f = req.app.get("u2f");
  const appid = u2f_common.extract_app_id(req);

  logger.debug("U2F register_request: headers=%s", JSON.stringify(req.headers));
  logger.info("U2F register_request: Starting registration of app %s", appid);
  u2f.startRegistration(appid, [])
    .then(function (registrationRequest: authdog.AuthenticationRequest) {
      logger.info("U2F register_request: Sending back registration request");
      req.session.auth_session.register_request = registrationRequest;
      res.status(200);
      res.json(registrationRequest);
    })
    .catch(function (err: Error) {
      logger.error("U2F register_request: %s", err);
      res.status(500);
      res.send("Unable to start registration request");
    });
}

function register(req: express.Request, res: express.Response) {
  const registrationRequest = objectPath.get(req, "session.auth_session.register_request");
  const challenge = objectPath.get(req, "session.auth_session.identity_check.challenge");

  if (!registrationRequest) {
    res.status(403);
    res.send();
    return;
  }

  if (!(registrationRequest && challenge == "u2f-register")) {
    res.status(403);
    res.send();
    return;
  }


  const user_data_storage = req.app.get("user data store");
  const u2f = req.app.get("u2f");
  const userid = req.session.auth_session.userid;
  const appid = u2f_common.extract_app_id(req);
  const logger = req.app.get("logger");

  logger.info("U2F register: Finishing registration");
  logger.debug("U2F register: register_request=%s", JSON.stringify(registrationRequest));
  logger.debug("U2F register: body=%s", JSON.stringify(req.body));

  u2f.finishRegistration(registrationRequest, req.body)
    .then(function (registrationStatus: authdog.Registration) {
      logger.info("U2F register: Store registration and reply");
      const meta = {
        keyHandle: registrationStatus.keyHandle,
        publicKey: registrationStatus.publicKey,
        certificate: registrationStatus.certificate
      };
      return user_data_storage.set_u2f_meta(userid, appid, meta);
    })
    .then(function () {
      objectPath.set(req, "session.auth_session.identity_check", undefined);
      res.status(204);
      res.send();
    })
    .catch(function (err: Error) {
      logger.error("U2F register: %s", err);
      res.status(500);
      res.send("Unable to register");
    });
}

export = {
  register_request: register_request,
  register: register
};
