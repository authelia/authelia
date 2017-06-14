
import request_ = require("request");
import assert = require("assert");
import speakeasy = require("speakeasy");
import BluebirdPromise = require("bluebird");
import util = require("util");
import sinon = require("sinon");
import Endpoints = require("../../src/server/endpoints");

const j = request_.jar();
const request: typeof request_ = <typeof request_>BluebirdPromise.promisifyAll(request_.defaults({ jar: j }));

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

const AUTHELIA_HOST = "nginx";
const DOMAIN = "test.local";
const PORT = 8080;

const HOME_URL = util.format("https://%s.%s:%d", "home", DOMAIN, PORT);
const SECRET_URL = util.format("https://%s.%s:%d", "secret", DOMAIN, PORT);
const SECRET1_URL = util.format("https://%s.%s:%d", "secret1", DOMAIN, PORT);
const SECRET2_URL = util.format("https://%s.%s:%d", "secret2", DOMAIN, PORT);
const MX1_URL = util.format("https://%s.%s:%d", "mx1.mail", DOMAIN, PORT);
const MX2_URL = util.format("https://%s.%s:%d", "mx2.mail", DOMAIN, PORT);
const BASE_AUTH_URL = util.format("https://%s.%s:%d", "auth", DOMAIN, PORT);

describe("test the server", function () {
  let home_page: string;
  let login_page: string;

  before(function () {
    const home_page_promise = getHomePage()
      .then(function (data) {
        home_page = data.body;
      });
    const login_page_promise = getLoginPage()
      .then(function (data) {
        login_page = data.body;
      });
    return BluebirdPromise.all([home_page_promise,
      login_page_promise]);
  });

  function str_contains(str: string, pattern: string) {
    return str.indexOf(pattern) != -1;
  }

  function home_page_contains(pattern: string) {
    return str_contains(home_page, pattern);
  }

  it("should serve a correct home page", function () {
    assert(home_page_contains(BASE_AUTH_URL + Endpoints.LOGOUT_GET + "?redirect=" + HOME_URL + "/"));
    assert(home_page_contains(HOME_URL + "/secret.html"));
    assert(home_page_contains(SECRET_URL + "/secret.html"));
    assert(home_page_contains(SECRET1_URL + "/secret.html"));
    assert(home_page_contains(SECRET2_URL + "/secret.html"));
    assert(home_page_contains(MX1_URL + "/secret.html"));
    assert(home_page_contains(MX2_URL + "/secret.html"));
  });

  it("should serve the login page", function (done) {
    getPromised(BASE_AUTH_URL + Endpoints.FIRST_FACTOR_GET + "?redirect=/")
      .then(function (data: request_.RequestResponse) {
        assert.equal(data.statusCode, 200);
        done();
      });
  });

  it("should serve the homepage", function (done) {
    getPromised(HOME_URL + "/")
      .then(function (data: request_.RequestResponse) {
        assert.equal(data.statusCode, 200);
        done();
      });
  });

  it("should redirect when logout", function (done) {
    getPromised(BASE_AUTH_URL + Endpoints.LOGOUT_GET + "?redirect=" + HOME_URL)
      .then(function (data: request_.RequestResponse) {
        assert.equal(data.statusCode, 200);
        assert.equal(data.body, home_page);
        done();
      });
  });

  it("should be redirected to the login page when accessing secret while not authenticated", function (done) {
    const url = HOME_URL + "/secret.html";
    getPromised(url)
      .then(function (data: request_.RequestResponse) {
        assert.equal(data.statusCode, 200);
        assert.equal(data.body, login_page);
        done();
      });
  });

  it.skip("should fail the first factor", function (done) {
    postPromised(BASE_AUTH_URL + Endpoints.FIRST_FACTOR_POST, {
      form: {
        username: "admin",
        password: "password",
      }
    })
      .then(function (data: request_.RequestResponse) {
        assert.equal(data.body, "Bad credentials");
        done();
      });
  });

  function login_as(username: string, password: string) {
    return postPromised(BASE_AUTH_URL + Endpoints.FIRST_FACTOR_POST, {
      form: {
        username: "john",
        password: "password",
      }
    })
      .then(function (data: request_.RequestResponse) {
        assert.equal(data.statusCode, 302);
        return BluebirdPromise.resolve();
      });
  }

  it("should succeed the first factor", function () {
    return login_as("john", "password");
  });

  describe("test ldap connection", function () {
    it("should not fail after inactivity", function () {
      const clock = sinon.useFakeTimers();
      return login_as("john", "password")
        .then(function () {
          clock.tick(3600000 * 24); // 24 hour
          return login_as("john", "password");
        })
        .then(function () {
          clock.restore();
          return BluebirdPromise.resolve();
        });
    });
  });
});

function getPromised(url: string) {
  return request.getAsync(url);
}

function postPromised(url: string, body: Object) {
  return request.postAsync(url, body);
}

function getHomePage(): BluebirdPromise<request_.RequestResponse> {
  return getPromised(HOME_URL + "/");
}

function getLoginPage(): BluebirdPromise<request_.RequestResponse> {
  return getPromised(BASE_AUTH_URL + Endpoints.FIRST_FACTOR_GET);
}
