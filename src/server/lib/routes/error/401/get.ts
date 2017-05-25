

import express = require("express");

export default function (req: express.Request, res: express.Response) {
    res.render("errors/401");
}
