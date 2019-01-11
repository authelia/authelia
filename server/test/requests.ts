
import BluebirdPromise = require("bluebird");
import request = require("request");
import assert = require("assert");
import Endpoints = require("../../shared/api");

declare module "request" {
  export interface RequestAPI<TRequest extends Request,
      TOptions extends CoreOptions,
      TUriUrlOptions> {
      getAsync(uri: string, options?: RequiredUriUrl): BluebirdPromise<RequestResponse>;
      getAsync(uri: string): BluebirdPromise<RequestResponse>;
      getAsync(options: RequiredUriUrl & CoreOptions): BluebirdPromise<RequestResponse>;

      postAsync(uri: string, options?: CoreOptions): BluebirdPromise<RequestResponse>;
      postAsync(uri: string): BluebirdPromise<RequestResponse>;
      postAsync(options: RequiredUriUrl & CoreOptions): BluebirdPromise<RequestResponse>;
  }
}

const requestAsync: typeof request = BluebirdPromise.promisifyAll(request) as typeof request;

export = function (port: number) {
  const PORT = port;
  const BASE_URL = "http://localhost:" + PORT;

  function execute_totp(jar: request.CookieJar, token: string) {
    return requestAsync.postAsync({
      url: BASE_URL + Endpoints.SECOND_FACTOR_TOTP_POST,
      jar: jar,
      form: {
        token: token
      }
    });
  }

  function execute_u2f_authentication(jar: request.CookieJar) {
    return requestAsync.getAsync({
      url: BASE_URL + Endpoints.SECOND_FACTOR_U2F_SIGN_REQUEST_GET,
      jar: jar
    })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        return requestAsync.postAsync({
          url: BASE_URL + Endpoints.SECOND_FACTOR_U2F_SIGN_POST,
          jar: jar,
          form: {
          }
        });
      });
  }

  function execute_verification(jar: request.CookieJar) {
    return requestAsync.getAsync({ url: BASE_URL + Endpoints.VERIFY_GET, jar: jar });
  }

  function execute_login(jar: request.CookieJar) {
    return requestAsync.getAsync({ url: BASE_URL + Endpoints.FIRST_FACTOR_GET, jar: jar });
  }

  function execute_first_factor(jar: request.CookieJar) {
    return requestAsync.postAsync({
      url: BASE_URL + Endpoints.FIRST_FACTOR_POST,
      jar: jar,
      form: {
        username: "test_ok",
        password: "password"
      }
    });
  }

  function execute_failing_first_factor(jar: request.CookieJar) {
    return requestAsync.postAsync({
      url: BASE_URL + Endpoints.FIRST_FACTOR_POST,
      jar: jar,
      form: {
        username: "test_nok",
        password: "password"
      }
    });
  }

  return {
    login: execute_login,
    verify: execute_verification,
    u2f_authentication: execute_u2f_authentication,
    first_factor: execute_first_factor,
    failing_first_factor: execute_failing_first_factor,
    totp: execute_totp,
  };
};

