
import util = require("util");
import express = require("express");

function extract_app_id(req: express.Request) {
  return util.format("https://%s", req.headers.host);
}

function extract_original_url(req: express.Request) {
  return util.format("https://%s%s", req.headers.host, req.headers["x-original-uri"]);
}

function extract_referrer(req: express.Request) {
  return req.headers.referrer;
}

function reply_with_internal_error(res: express.Response, msg: string) {
  res.status(500);
  res.send(msg);
}

function reply_with_missing_registration(res: express.Response) {
  res.status(401);
  res.send("Please register before authenticate");
}

function reply_with_unauthorized(res: express.Response) {
  res.status(401);
  res.send();
}

export = {
  extract_app_id: extract_app_id,
  extract_original_url: extract_original_url,
  extract_referrer: extract_referrer,
  reply_with_internal_error: reply_with_internal_error,
  reply_with_missing_registration: reply_with_missing_registration,
  reply_with_unauthorized: reply_with_unauthorized
};