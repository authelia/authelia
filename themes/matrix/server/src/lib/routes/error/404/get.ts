
import BluebirdPromise = require("bluebird");
import express = require("express");

export default function (req: express.Request, res: express.Response): BluebirdPromise<void> {
    res.render("errors/404");
    return BluebirdPromise.resolve();
}