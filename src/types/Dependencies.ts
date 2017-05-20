import winston = require("winston");
import speakeasy = require("speakeasy");
import nodemailer = require("nodemailer");
import session = require("express-session");
import nedb = require("nedb");
import ldapjs = require("ldapjs");

export type Nodemailer = typeof nodemailer;
export type Speakeasy = typeof speakeasy;
export type Winston = typeof winston;
export type Session = typeof session;
export type Nedb = typeof nedb;

export interface GlobalDependencies {
    u2f: object;
    nodemailer: Nodemailer;
    ldapjs: object;
    session: Session;
    winston: Winston;
    speakeasy: Speakeasy;
    nedb: Nedb;
}