/*
 * This is just client script to test the fake API.
 */

const DuoApi = require("@duosecurity/duo_api");

process.env["NODE_TLS_REJECT_UNAUTHORIZED"] = 0;

const client = new DuoApi.Client("ABCDEFG", "SECRET", "duo.example.com");
client.jsonApiCall("POST", "/auth/v2/auth", { username: 'john', factor: "push", device: "auto" }, console.log);