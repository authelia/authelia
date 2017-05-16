import * as winston from "winston";

export interface GlobalDependencies {
    u2f: object;
    nodemailer: any;
    ldapjs: object;
    session: any;
    winston: winston.Winston;
    speakeasy: object;
    nedb: any;
}