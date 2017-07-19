import SessionConfigurationBuilder from "../../../src/server/lib/SessionConfigurationBuilder";
import { AppConfiguration } from "../../../src/types/Configuration";
import { GlobalDependencies } from "../../../src/types/Dependencies";
import ExpressSession = require("express-session");
import ConnectRedis = require("connect-redis");
import Sinon = require("sinon");
import Assert = require("assert");

describe("test session configuration builder", function () {
    it("should return session options without redis options", function () {
        const configuration: AppConfiguration = {
            access_control: {
                default: [],
                users: {},
                groups: {}
            },
            ldap: {
                url: "ldap://ldap",
                base_dn: "dc=example,dc=com",
                user: "user",
                password: "password"
            },
            logs_level: "debug",
            notifier: {
                filesystem: {
                    filename: "/test"
                }
            },
            port: 8080,
            session: {
                domain: "example.com",
                expiration: 3600,
                secret: "secret"
            },
            store_in_memory: true
        };

        const deps: GlobalDependencies = {
            ConnectRedis: Sinon.spy() as any,
            ldapjs: Sinon.spy() as any,
            nedb: Sinon.spy() as any,
            nodemailer: Sinon.spy() as any,
            session: Sinon.spy() as any,
            speakeasy: Sinon.spy() as any,
            u2f: Sinon.spy() as any,
            winston: Sinon.spy() as any,
            dovehash: Sinon.spy() as any
        };

        const options = SessionConfigurationBuilder.build(configuration, deps);

        const expectedOptions = {
            secret: "secret",
            resave: false,
            saveUninitialized: true,
            cookie: {
                secure: false,
                maxAge: 3600,
                domain: "example.com"
            }
        };

        Assert.deepEqual(expectedOptions, options);
    });

    it("should return session options with redis options", function () {
        const configuration: AppConfiguration = {
            access_control: {
                default: [],
                users: {},
                groups: {}
            },
            ldap: {
                url: "ldap://ldap",
                base_dn: "dc=example,dc=com",
                user: "user",
                password: "password"
            },
            logs_level: "debug",
            notifier: {
                filesystem: {
                    filename: "/test"
                }
            },
            port: 8080,
            session: {
                domain: "example.com",
                expiration: 3600,
                secret: "secret",
                redis: {
                    host: "redis.example.com",
                    port: 6379
                }
            },
            store_in_memory: true
        };

        const RedisStoreMock = Sinon.spy();

        const deps: GlobalDependencies = {
            ConnectRedis: Sinon.stub().returns(RedisStoreMock) as any,
            ldapjs: Sinon.spy() as any,
            nedb: Sinon.spy() as any,
            nodemailer: Sinon.spy() as any,
            session: Sinon.spy() as any,
            speakeasy: Sinon.spy() as any,
            u2f: Sinon.spy() as any,
            winston: Sinon.spy() as any,
            dovehash: Sinon.spy() as any
        };

        const options = SessionConfigurationBuilder.build(configuration, deps);

        const expectedOptions: ExpressSession.SessionOptions = {
            secret: "secret",
            resave: false,
            saveUninitialized: true,
            cookie: {
                secure: false,
                maxAge: 3600,
                domain: "example.com"
            },
            store: Sinon.match.object as any
        };

        Assert((deps.ConnectRedis as Sinon.SinonStub).calledWith(deps.session));
        Assert.equal(options.secret, expectedOptions.secret);
        Assert.equal(options.resave, expectedOptions.resave);
        Assert.equal(options.saveUninitialized, expectedOptions.saveUninitialized);
        Assert.deepEqual(options.cookie, expectedOptions.cookie);
        Assert(options.store != undefined);
    });
});