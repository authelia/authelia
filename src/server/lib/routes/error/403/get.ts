
import BluebirdPromise = require("bluebird");
import express = require("express");

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    res.render("errors/403");
    return BluebirdPromise.resolve();
}