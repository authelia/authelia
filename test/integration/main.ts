
process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

import Request = require("request");
import Assert = require("assert");
import BluebirdPromise = require("bluebird");
import Util = require("util");
import Redis = require("redis");
import Endpoints = require("../../src/server/endpoints");

const RequestAsync = BluebirdPromise.promisifyAll(Request) as typeof Request;

const DOMAIN = "test.local";
const PORT = 8080;

const HOME_URL = Util.format("https://%s.%s:%d", "home", DOMAIN, PORT);
const SECRET_URL = Util.format("https://%s.%s:%d", "secret", DOMAIN, PORT);
const SECRET1_URL = Util.format("https://%s.%s:%d", "secret1", DOMAIN, PORT);
const SECRET2_URL = Util.format("https://%s.%s:%d", "secret2", DOMAIN, PORT);
const MX1_URL = Util.format("https://%s.%s:%d", "mx1.mail", DOMAIN, PORT);
const MX2_URL = Util.format("https://%s.%s:%d", "mx2.mail", DOMAIN, PORT);

const BASE_AUTH_URL = Util.format("https://%s.%s:%d", "auth", DOMAIN, PORT);
const FIRST_FACTOR_URL = Util.format("%s/api/firstfactor", BASE_AUTH_URL);
const LOGOUT_URL = Util.format("%s/logout", BASE_AUTH_URL);


const redisOptions = {
    host: "redis",
    port: 6379
};


describe("integration tests", function () {
    let redisClient: Redis.RedisClient;

    before(function () {
        redisClient = Redis.createClient(redisOptions);
    });

    function str_contains(str: string, pattern: string) {
        return str.indexOf(pattern) != -1;
    }

    function test_homepage_is_correct(body: string) {
        Assert(str_contains(body, BASE_AUTH_URL + Endpoints.LOGOUT_GET + "?redirect=" + HOME_URL + "/"));
        Assert(str_contains(body, HOME_URL + "/secret.html"));
        Assert(str_contains(body, SECRET_URL + "/secret.html"));
        Assert(str_contains(body, SECRET1_URL + "/secret.html"));
        Assert(str_contains(body, SECRET2_URL + "/secret.html"));
        Assert(str_contains(body, MX1_URL + "/secret.html"));
        Assert(str_contains(body, MX2_URL + "/secret.html"));
        Assert(str_contains(body, "Access the secret"));
    }

    it("should access the home page", function () {
        return RequestAsync.getAsync(HOME_URL)
            .then(function (response: Request.RequestResponse) {
                Assert.equal(200, response.statusCode);
                test_homepage_is_correct(response.body);
            });
    });

    it("should access the authentication page", function () {
        return RequestAsync.getAsync(BASE_AUTH_URL)
            .then(function (response: Request.RequestResponse) {
                Assert.equal(200, response.statusCode);
                Assert(response.body.indexOf("Sign in") > -1);
            });
    });

    it("should fail first factor when wrong credentials are provided", function () {
        return RequestAsync.postAsync(FIRST_FACTOR_URL, {
            json: true,
            body: {
                username: "john",
                password: "wrong password"
            }
        })
            .then(function (response: Request.RequestResponse) {
                Assert.equal(401, response.statusCode);
            });
    });

    it("should redirect when correct credentials are provided during first factor", function () {
        return RequestAsync.postAsync(FIRST_FACTOR_URL, {
            json: true,
            body: {
                username: "john",
                password: "password"
            }
        })
            .then(function (response: Request.RequestResponse) {
                Assert.equal(302, response.statusCode);
            });
    });

    it("should have registered four sessions in redis", function (done) {
        redisClient.dbsize(function (err: Error, count: number) {
            Assert.equal(3, count);
            done();
        });
    });

    it("should redirect to home page when logout is called", function () {
        return RequestAsync.getAsync(Util.format("%s?redirect=%s", LOGOUT_URL, HOME_URL))
            .then(function (response: Request.RequestResponse) {
                Assert.equal(200, response.statusCode);
                Assert(response.body.indexOf("Access the secret") > -1);
            });
    });
});