#! /usr/bin/env node

import Server from "./lib/Server";
import { GlobalDependencies } from "../types/Dependencies";
import YAML = require("yamljs");

const configurationFilepath = process.argv[2];
if (!configurationFilepath) {
  console.log("No config file has been provided.");
  console.log("Usage: authelia <config>");
  process.exit(0);
}

const yamlContent = YAML.load(configurationFilepath);

const deps: GlobalDependencies = {
  u2f: require("u2f"),
  ldapjs: require("ldapjs"),
  session: require("express-session"),
  winston: require("winston"),
  speakeasy: require("speakeasy"),
  nedb: require("nedb"),
  ConnectRedis: require("connect-redis"),
  Redis: require("redis")
};

const server = new Server(deps);
server.start(yamlContent, deps);
