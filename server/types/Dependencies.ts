import winston = require("winston");
import speakeasy = require("speakeasy");
import nodemailer = require("nodemailer");
import session = require("express-session");
import nedb = require("nedb");
import ldapjs = require("ldapjs");
import u2f = require("u2f");
import RedisSession = require("connect-redis");
import dovehash = require("dovehash");

export type Dovehash = typeof dovehash;
export type Speakeasy = typeof speakeasy;
export type Winston = typeof winston;
export type Session = typeof session;
export type Nedb = typeof nedb;
export type Ldapjs = typeof ldapjs;
export type U2f = typeof u2f;
export type ConnectRedis = typeof RedisSession;

export interface GlobalDependencies {
    u2f: U2f;
    dovehash: Dovehash;
    ldapjs: Ldapjs;
    session: Session;
    ConnectRedis: ConnectRedis;
    winston: Winston;
    speakeasy: Speakeasy;
    nedb: Nedb;
}