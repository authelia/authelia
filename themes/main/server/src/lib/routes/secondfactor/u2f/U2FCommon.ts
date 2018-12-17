
import util = require("util");
import express = require("express");

function extract_app_id(req: express.Request): string {
  return util.format("https://%s", req.headers.host);
}

export = {
  extract_app_id: extract_app_id
};