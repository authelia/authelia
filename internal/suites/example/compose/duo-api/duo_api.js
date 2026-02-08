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
const app = express();
const port = 3000;

app.use(express.json());
app.set("trust proxy", true);

// Auth API
let permission = "allow";

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

app.post("/auth/v2/auth", (req, res) => {
    setTimeout(() => {
        let response;
        if (permission == "allow") {
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

// PreAuth API
let preauth = {
    result: "allow",
    status_msg: "Allowing unknown user",
};

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
