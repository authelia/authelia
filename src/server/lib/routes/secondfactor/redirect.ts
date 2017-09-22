
import express = require("express");
import objectPath = require("object-path");
import winston = require("winston");
import Endpoints = require("../../../endpoints");
import { ServerVariablesHandler } from "../../ServerVariablesHandler";
import AuthenticationSession = require("../../AuthenticationSession");

export default function (req: express.Request, res: express.Response) {
    const authSession = AuthenticationSession.get(req);
    const redirectUrl = req.query.redirect || Endpoints.FIRST_FACTOR_GET;
    res.json({
        redirection_url: redirectUrl
    });
}