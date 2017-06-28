
import Request = require("request");
import Assert = require("assert");
import Speakeasy = require("speakeasy");
import BluebirdPromise = require("bluebird");
import Util = require("util");
import Sinon = require("sinon");
import Endpoints = require("../../src/server/endpoints");

const EXEC_PATH = "./dist/src/server/index.js";
const CONFIG_PATH = "./test/integration/config.yml";
const j = Request.jar();
const request: typeof Request = <typeof Request>BluebirdPromise.promisifyAll(Request.defaults({ jar: j }));

process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";

const DOMAIN = "test.local";
const PORT = 8080;

const HOME_URL = Util.format("https://%s.%s:%d", "home", DOMAIN, PORT);
const SECRET_URL = Util.format("https://%s.%s:%d", "secret", DOMAIN, PORT);
const SECRET1_URL = Util.format("https://%s.%s:%d", "secret1", DOMAIN, PORT);
const SECRET2_URL = Util.format("https://%s.%s:%d", "secret2", DOMAIN, PORT);
const MX1_URL = Util.format("https://%s.%s:%d", "mx1.mail", DOMAIN, PORT);
const MX2_URL = Util.format("https://%s.%s:%d", "mx2.mail", DOMAIN, PORT);
const BASE_AUTH_URL = Util.format("https://%s.%s:%d", "auth", DOMAIN, PORT);

function waitFor(ms: number): BluebirdPromise<{}> {
  return new BluebirdPromise(function (resolve, reject) {
    setTimeout(function () {
      resolve();
    }, ms);
  });
}

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

  after(function () {
  });

  function str_contains(str: string, pattern: string) {
    return str.indexOf(pattern) != -1;
  }

  function home_page_contains(pattern: string) {
    return str_contains(home_page, pattern);
  }

  it("should serve a correct home page", function () {
    Assert(home_page_contains(BASE_AUTH_URL + Endpoints.LOGOUT_GET + "?redirect=" + HOME_URL + "/"));
    Assert(home_page_contains(HOME_URL + "/secret.html"));
    Assert(home_page_contains(SECRET_URL + "/secret.html"));
    Assert(home_page_contains(SECRET1_URL + "/secret.html"));
    Assert(home_page_contains(SECRET2_URL + "/secret.html"));
    Assert(home_page_contains(MX1_URL + "/secret.html"));
    Assert(home_page_contains(MX2_URL + "/secret.html"));
  });

  it("should serve the login page", function () {
    return getPromised(BASE_AUTH_URL + Endpoints.FIRST_FACTOR_GET)
      .then(function (data: Request.RequestResponse) {
        Assert.equal(data.statusCode, 200);
      });
  });

  it("should serve the homepage", function () {
    return getPromised(HOME_URL + "/")
      .then(function (data: Request.RequestResponse) {
        Assert.equal(data.statusCode, 200);
      });
  });

  it("should redirect when logout", function () {
    return getPromised(BASE_AUTH_URL + Endpoints.LOGOUT_GET + "?redirect=" + HOME_URL)
      .then(function (data: Request.RequestResponse) {
        Assert.equal(data.statusCode, 200);
        Assert.equal(data.body, home_page);
      });
  });

  it("should be redirected to the login page when accessing secret while not authenticated", function () {
    return getPromised(HOME_URL + "/secret.html")
      .then(function (data: Request.RequestResponse) {
        Assert.equal(data.statusCode, 200);
        Assert.equal(data.body, login_page);
      });
  });

  it.skip("should fail the first factor", function () {
    return postPromised(BASE_AUTH_URL + Endpoints.FIRST_FACTOR_POST, {
      form: {
        username: "admin",
        password: "password",
      }
    })
      .then(function (data: Request.RequestResponse) {
        Assert.equal(data.body, "Bad credentials");
      });
  });

  function login_as(username: string, password: string) {
    return postPromised(BASE_AUTH_URL + Endpoints.FIRST_FACTOR_POST, {
      form: {
        username: "john",
        password: "password",
      }
    })
      .then(function (data: Request.RequestResponse) {
        Assert.equal(data.statusCode, 302);
        return BluebirdPromise.resolve();
      });
  }

  it("should succeed the first factor", function () {
    return login_as("john", "password");
  });

  describe("test ldap connection", function () {
    it("should not fail after inactivity", function () {
      const clock = Sinon.useFakeTimers();
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

function getHomePage(): BluebirdPromise<Request.RequestResponse> {
  return getPromised(HOME_URL + "/");
}

function getLoginPage(): BluebirdPromise<Request.RequestResponse> {
  return getPromised(BASE_AUTH_URL + Endpoints.FIRST_FACTOR_GET);
}
