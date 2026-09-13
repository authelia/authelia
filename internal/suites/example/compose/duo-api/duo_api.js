// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

/*
 * This is a script to fake the Duo API for push notifications.
 *
 * For Auth API access is allowed by default but one can change the
 * behavior at runtime by POSTing to /allow or /deny. Then the /auth/v2/auth
 * endpoint will act accordingly.
 *
 * For PreAuth API device selection is bypassed by default but one can
 * change the behavior at runtime by POSTing to /preauth using the desired
 * result parameters (and devices). Then the /auth/v2/preauth endpoint
 * will act accordingly.
 */

const express = require("express");
const crypto = require("crypto");

const port = 3000;

const INTEGRATION_KEY = "ABCDEFGHIJKL";
const SECRET_KEY = "abcdefghijklmnopqrstuvwxyz123456789";

const app = express();

app.use(express.json());
app.set("trust proxy", true);

app.post("/allow", (req, res) => {
    permission = "allow";
    console.log("auth set allowed!");
    res.send("ALLOWED");
});

app.post("/deny", (req, res) => {
    permission = "deny";
    console.log("auth set denied!");
    res.send("DENIED");
});

app.get("/auth/v2/ping", (req, res) => {
    res.status(200).json(status());
});

app.get("/auth/v2/check", auth, (req, res) => {
    res.status(200).json(status());
});

app.post("/auth/v2/auth", (req, res) => {
    setTimeout(() => {
        let response;
        if (permission === "allow") {
            response = {
                response: {
                    result: "allow",
                    status: "allow",
                    status_msg: "The user allowed access.",
                },
                stat: "OK",
            };
        } else {
            response = {
                response: {
                    result: "deny",
                    status: "deny",
                    status_msg: "The user denied access.",
                },
                stat: "OK",
            };
        }
        res.json(response);
        console.log("Auth API responded with %s", permission);
    }, 2000);
});

app.post("/preauth", (req, res) => {
    preauth = req.body;
    console.log("set result to: %s", preauth);
    res.json(preauth);
});

app.post("/auth/v2/preauth", (req, res) => {
    setTimeout(() => {
        let response;
        response = {
            response: preauth,
            stat: "OK",
        };

        res.json(response);
        console.log("PreAuth API responded with %s", preauth);
    }, 2000);
});

app.listen(port, () => console.log(`Duo API listening on port ${port}!`));

// Auth API
let permission = "allow";

// PreAuth API
let preauth = {
    result: "allow",
    status_msg: "Allowing unknown user",
};

// The signals we want to handle
// NOTE: although it is tempting, the SIGKILL signal (9) cannot be intercepted and handled
var signals = {
    SIGHUP: 1,
    SIGINT: 2,
    SIGTERM: 15,
};

// Create a listener for each of the signals that we want to handle
Object.keys(signals).forEach((signal) => {
    process.on(signal, () => {
        console.log(`process received a ${signal} signal`);
        process.exit(128 + signals[signal]);
    });
});

/*
 * Duo signs requests with a HMAC-SHA512 of the canonical request which is transmitted via the Basic scheme of the
 * Authorization header as the password, with the integration key as the username. See the signature documentation at
 * https://duo.com/docs/authapi#authentication.
 */
function canonicalize(method, host, uri, params, date) {
    return [date, method.toUpperCase(), host.toLowerCase(), uri, params].join("\n");
}

function sign(method, host, uri, params, date) {
    return crypto.createHmac("sha512", SECRET_KEY).update(canonicalize(method, host, uri, params, date)).digest("hex");
}

function unauthorized(res) {
    res.set("WWW-Authenticate", 'Basic realm="Restricted"');

    return res.status(401).send("Authentication required");
}

function auth(req, res, next) {
    const header = req.headers.authorization;

    if (!header || !header.startsWith("Basic ")) {
        return unauthorized(res);
    }

    let integration_key = "", signature = "";

    try {
        const [, raw] = header.split(" ");
        const [i, s] = Buffer.from(raw, "base64").toString("utf8").split(":");

        integration_key = i || "";
        signature = s || "";
    } catch {
        return res.status(400).send("Bad Request");
    }

    if (integration_key !== INTEGRATION_KEY) {
        return unauthorized(res);
    }

    // The canonical request percent encodes spaces rather than encoding them as a plus.
    const [, query = ""] = req.originalUrl.split("?");
    const params = query.replace(/\+/g, "%20");

    const expected = sign(req.method, req.headers.host || "", req.path, params, req.headers.date || "");

    if (signature.length !== expected.length || !crypto.timingSafeEqual(Buffer.from(signature), Buffer.from(expected))) {
        return unauthorized(res);
    }

    return next();
}

function status() {
    return {stat: "OK", response: {time: Math.floor(Date.now() / 1000)}}
}
