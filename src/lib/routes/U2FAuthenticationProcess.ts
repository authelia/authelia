
import u2f_register_handler = require("./U2FRegistration");
import objectPath = require("object-path");
import u2f_common = require("./u2f_common");
import BluebirdPromise = require("bluebird");
import express = require("express");
import authdog = require("../../types/authdog");
import UserDataStore, { U2FMetaDocument } from "../UserDataStore";


function retrieve_u2f_meta(req: express.Request, userDataStore: UserDataStore) {
  const userid = req.session.auth_session.userid;
  const appid = u2f_common.extract_app_id(req);
  return userDataStore.get_u2f_meta(userid, appid);
}


function sign_request(req: express.Request, res: express.Response) {
  const logger = req.app.get("logger");
  const userDataStore = req.app.get("user data store");

  retrieve_u2f_meta(req, userDataStore)
    .then(function (doc: U2FMetaDocument) {
      if (!doc) {
        u2f_common.reply_with_missing_registration(res);
        return;
      }

      const u2f = req.app.get("u2f");
      const meta = doc.meta;
      const appid = u2f_common.extract_app_id(req);
      logger.info("U2F sign_request: Start authentication to app %s", appid);
      return u2f.startAuthentication(appid, [meta]);
    })
    .then(function (authRequest: authdog.AuthenticationRequest) {
      logger.info("U2F sign_request: Store authentication request and reply");
      req.session.auth_session.sign_request = authRequest;
      res.status(200);
      res.json(authRequest);
    })
    .catch(function (err: Error) {
      logger.info("U2F sign_request: %s", err);
      res.status(500);
      res.send();
    });
}


function sign(req: express.Request, res: express.Response) {
  if (!objectPath.has(req, "session.auth_session.sign_request")) {
    u2f_common.reply_with_unauthorized(res);
    return;
  }

  const logger = req.app.get("logger");
  const userDataStore = req.app.get("user data store");

  retrieve_u2f_meta(req, userDataStore)
    .then(function (doc: U2FMetaDocument) {
      const appid = u2f_common.extract_app_id(req);
      const u2f = req.app.get("u2f");
      const authRequest = req.session.auth_session.sign_request;
      const meta = doc.meta;
      logger.info("U2F sign: Finish authentication");
      return u2f.finishAuthentication(authRequest, req.body, [meta]);
    })
    .then(function (authenticationStatus: authdog.Authentication) {
      logger.info("U2F sign: Authentication successful");
      req.session.auth_session.second_factor = true;
      res.status(204);
      res.send();
    })
    .catch(function (err: Error) {
      logger.error("U2F sign: %s", err);
      res.status(500);
      res.send();
    });
}


export = {
  sign_request: sign_request,
  sign: sign
};
