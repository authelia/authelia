
import express = require("express");
import objectPath = require("object-path");
import winston = require("winston");
import Endpoints = require("../../../endpoints");
import ServerVariables = require("../../ServerVariables");
import AuthenticationSession = require("../../AuthenticationSession");

export default function (req: express.Request, res: express.Response) {
    const authSession = AuthenticationSession.get(req);
    const redirectUrl = authSession.redirect || Endpoints.FIRST_FACTOR_GET;
    res.json({
        redirection_url: redirectUrl
    });
}