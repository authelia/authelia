#! /usr/bin/env node

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

import Server from "./lib/Server";
const YAML = require("yamljs");

const config_path = process.argv[2];
if (!config_path) {
  console.log("No config file has been provided.");
  console.log("Usage: authelia <config>");
  process.exit(0);
}

console.log("Parse configuration file: %s", config_path);

const yaml_config = YAML.load(config_path);

const deps = {
  u2f: require("u2f"),
  nodemailer: require("nodemailer"),
  ldapjs: require("ldapjs"),
  session: require("express-session"),
  winston: require("winston"),
  speakeasy: require("speakeasy"),
  nedb: require("nedb")
};

const server = new Server();
server.start(yaml_config, deps)
.then(() => {
  console.log("The server is started!");
});
