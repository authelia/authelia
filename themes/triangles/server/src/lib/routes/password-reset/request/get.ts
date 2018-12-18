
import express = require("express");
import BluebirdPromise = require("bluebird");
import objectPath = require("object-path");
import exceptions = require("../../../Exceptions");

import Constants = require("./../constants");

const TEMPLATE_NAME = "password-reset-request";

export default function (req: express.Request, res: express.Response) {
    res.render(TEMPLATE_NAME);
}