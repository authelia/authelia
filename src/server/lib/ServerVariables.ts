
import winston = require("winston");
import { LdapClient } from "./LdapClient";
import { TOTPValidator } from "./TOTPValidator";
import { TOTPGenerator } from "./TOTPGenerator";
import U2F = require("u2f");
import UserDataStore from "./UserDataStore";
import { INotifier } from "./notifiers/INotifier";
import { AuthenticationRegulator } from "./AuthenticationRegulator";
import Configuration = require("../../types/Configuration");
import { AccessController } from "./access_control/AccessController";
import { NotifierFactory } from "./notifiers/NotifierFactory";

import { GlobalDependencies } from "../../types/Dependencies";

import express = require("express");

export const VARIABLES_KEY = "authelia-variables";

export interface ServerVariables {
    logger: typeof winston;
    ldap: LdapClient;
    totpValidator: TOTPValidator;
    totpGenerator: TOTPGenerator;
    u2f: typeof U2F;
    userDataStore: UserDataStore;
    notifier: INotifier;
    regulator: AuthenticationRegulator;
    config: Configuration.AppConfiguration;
    accessController: AccessController;
}


export function fill(app: express.Application, config: Configuration.AppConfiguration, deps: GlobalDependencies) {
    const five_minutes = 5 * 60;
    const datastore_options = {
        directory: config.store_directory,
        inMemory: config.store_in_memory
    };

    const userDataStore = new UserDataStore(datastore_options, deps.nedb);
    const regulator = new AuthenticationRegulator(userDataStore, five_minutes);
    const notifier = NotifierFactory.build(config.notifier, deps.nodemailer);
    const ldap = new LdapClient(config.ldap, deps.ldapjs, deps.winston);
    const accessController = new AccessController(config.access_control, deps.winston);
    const totpValidator = new TOTPValidator(deps.speakeasy);
    const totpGenerator = new TOTPGenerator(deps.speakeasy);

    const variables: ServerVariables = {
        accessController: accessController,
        config: config,
        ldap: ldap,
        logger: deps.winston,
        notifier: notifier,
        regulator: regulator,
        totpGenerator: totpGenerator,
        totpValidator: totpValidator,
        u2f: deps.u2f,
        userDataStore: userDataStore
    };

    app.set(VARIABLES_KEY, variables);
}

export function getLogger(app: express.Application): typeof winston {
    return (app.get(VARIABLES_KEY) as ServerVariables).logger;
}

export function getUserDataStore(app: express.Application): UserDataStore {
    return (app.get(VARIABLES_KEY) as ServerVariables).userDataStore;
}

export function getNotifier(app: express.Application): INotifier {
    return (app.get(VARIABLES_KEY) as ServerVariables).notifier;
}

export function getLdapClient(app: express.Application): LdapClient {
    return (app.get(VARIABLES_KEY) as ServerVariables).ldap;
}

export function getConfiguration(app: express.Application): Configuration.AppConfiguration {
    return (app.get(VARIABLES_KEY) as ServerVariables).config;
}

export function getAuthenticationRegulator(app: express.Application): AuthenticationRegulator {
    return (app.get(VARIABLES_KEY) as ServerVariables).regulator;
}

export function getAccessController(app: express.Application): AccessController {
    return (app.get(VARIABLES_KEY) as ServerVariables).accessController;
}

export function getTOTPGenerator(app: express.Application): TOTPGenerator {
    return (app.get(VARIABLES_KEY) as ServerVariables).totpGenerator;
}

export function getTOTPValidator(app: express.Application): TOTPValidator {
    return (app.get(VARIABLES_KEY) as ServerVariables).totpValidator;
}

export function getU2F(app: express.Application): typeof U2F {
    return (app.get(VARIABLES_KEY) as ServerVariables).u2f;
}
