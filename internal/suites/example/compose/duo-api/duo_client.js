// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

/*
 * This is just client script to test the fake API.
 */

const DuoApi = require("@duosecurity/duo_api");

const client = new DuoApi.Client("ABCDEFGHIJKL", "abcdefghijklmnopqrstuvwxyz123456789", "duo.example.com");

process.env["NODE_TLS_REJECT_UNAUTHORIZED"] = 0;

console.log("Testing Auth API first");

client.jsonApiCall(
    "POST",
    "/auth/v2/auth",
    { username: "john", factor: "push", device: "auto" },
    console.log,
);

console.log("Testing PreAuth API second");

client.jsonApiCall(
    "POST",
    "/auth/v2/preauth",
    { username: "john" },
    console.log,
);
