
import express = require("express");

const TEMPLATE_NAME = "password-reset-request";

export default function (req: express.Request, res: express.Response) {
    res.render(TEMPLATE_NAME);
}