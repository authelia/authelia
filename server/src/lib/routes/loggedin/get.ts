import Express = require("express");
import Endpoints = require("../../../../../shared/api");

export default function(req: Express.Request, res: Express.Response) {
  res.render("already-logged-in", {
    logout_endpoint: Endpoints.LOGOUT_GET
  });
}