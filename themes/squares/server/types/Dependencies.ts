import winston = require("winston");
import speakeasy = require("speakeasy");
import nodemailer = require("nodemailer");
import session = require("express-session");
import nedb = require("nedb");
import ldapjs = require("ldapjs");
import u2f = require("u2f");
import RedisSession = require("connect-redis");
import Redis = require("redis");

export type Speakeasy = typeof speakeasy;
export type Winston = typeof winston;
export type Session = typeof session;
export type Nedb = typeof nedb;
export type Ldapjs = typeof ldapjs;
export type U2f = typeof u2f;
export type ConnectRedis = typeof RedisSession;
export type Redis = typeof Redis;

export interface GlobalDependencies {
    u2f: U2f;
    ldapjs: Ldapjs;
    session: Session;
    Redis: Redis;
    ConnectRedis: ConnectRedis;
    winston: Winston;
    speakeasy: Speakeasy;
    nedb: Nedb;
}