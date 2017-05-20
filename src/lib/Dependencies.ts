import * as winston from "winston";
import nodemailer = require("nodemailer");

export interface Nodemailer {
    createTransport: (options?: any, defaults?: Object) => nodemailer.Transporter;
}

export interface GlobalDependencies {
    u2f: object;
    nodemailer: Nodemailer;
    ldapjs: object;
    session: any;
    winston: winston.Winston;
    speakeasy: object;
    nedb: any;
}

export type NodemailerDependencies = Nodemailer;

export interface NotifierDependencies {
    nodemailer: Nodemailer;
}