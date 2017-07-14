#! /usr/bin/env node

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

import Server from "./lib/Server";
import { GlobalDependencies } from "../types/Dependencies";
const YAML = require("yamljs");

const configurationFilepath = process.argv[2];
if (!configurationFilepath) {
  console.log("No config file has been provided.");
  console.log("Usage: authelia <config>");
  process.exit(0);
}

console.log("Parse configuration file: %s", configurationFilepath);

const yamlContent = YAML.load(configurationFilepath);

const deps: GlobalDependencies = {
  u2f: require("u2f"),
  nodemailer: require("nodemailer"),
  ldapjs: require("ldapjs"),
  session: require("express-session"),
  winston: require("winston"),
  speakeasy: require("speakeasy"),
  nedb: require("nedb"),
  ConnectRedis: require("connect-redis")
};

const server = new Server();
server.start(yamlContent, deps)
.then(() => {
  console.log("The server is started!");
});
