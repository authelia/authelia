
import BluebirdPromise = require("bluebird");
import request = require("request");
import assert = require("assert");
import express = require("express");
import nodemailer = require("nodemailer");

import NodemailerMock = require("./mocks/nodemailer");

const requestAsync = BluebirdPromise.promisifyAll(request) as request.RequestAsync;

export = function (port: number) {
  const PORT = port;
  const BASE_URL = "http://localhost:" + PORT;

  function execute_reset_password(jar: request.CookieJar, transporter: NodemailerMock.NodemailerTransporterMock, user: string, new_password: string) {
    return requestAsync.postAsync({
      url: BASE_URL + "/reset-password",
      jar: jar,
      form: { userid: user }
    })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 204);
        const html_content = transporter.sendMail.getCall(0).args[0].html;
        const regexp = /identity_token=([a-zA-Z0-9]+)/;
        const token = regexp.exec(html_content)[1];
        // console.log(html_content, token);
        return requestAsync.getAsync({
          url: BASE_URL + "/reset-password?identity_token=" + token,
          jar: jar
        });
      })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        return requestAsync.postAsync({
          url: BASE_URL + "/new-password",
          jar: jar,
          form: {
            password: new_password
          }
        });
      });
  }

  function execute_register_totp(jar: request.CookieJar, transporter: NodemailerMock.NodemailerTransporterMock) {
    return requestAsync.postAsync({
      url: BASE_URL + "/totp-register",
      jar: jar
    })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 204);
        const html_content = transporter.sendMail.getCall(0).args[0].html;
        const regexp = /identity_token=([a-zA-Z0-9]+)/;
        const token = regexp.exec(html_content)[1];
        // console.log(html_content, token);
        return requestAsync.getAsync({
          url: BASE_URL + "/totp-register?identity_token=" + token,
          jar: jar
        });
      })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        return requestAsync.postAsync({
          url: BASE_URL + "/new-totp-secret",
          jar: jar,
        });
      })
      .then(function (res: request.RequestResponse) {
        console.log(res.statusCode);
        console.log(res.body);
        assert.equal(res.statusCode, 200);
        return Promise.resolve(res.body);
      });
  }

  function execute_totp(jar: request.CookieJar, token: string) {
    return requestAsync.postAsync({
      url: BASE_URL + "/2ndfactor/totp",
      jar: jar,
      form: {
        token: token
      }
    });
  }

  function execute_u2f_authentication(jar: request.CookieJar) {
    return requestAsync.getAsync({
      url: BASE_URL + "/2ndfactor/u2f/sign_request",
      jar: jar
    })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        return requestAsync.postAsync({
          url: BASE_URL + "/2ndfactor/u2f/sign",
          jar: jar,
          form: {
          }
        });
      });
  }

  function execute_verification(jar: request.CookieJar) {
    return requestAsync.getAsync({ url: BASE_URL + "/verify", jar: jar });
  }

  function execute_login(jar: request.CookieJar) {
    return requestAsync.getAsync({ url: BASE_URL + "/login", jar: jar });
  }

  function execute_u2f_registration(jar: request.CookieJar, transporter: NodemailerMock.NodemailerTransporterMock) {
    return requestAsync.postAsync({
      url: BASE_URL + "/u2f-register",
      jar: jar
    })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 204);
        const html_content = transporter.sendMail.getCall(0).args[0].html;
        const regexp = /identity_token=([a-zA-Z0-9]+)/;
        const token = regexp.exec(html_content)[1];
        // console.log(html_content, token);
        return requestAsync.getAsync({
          url: BASE_URL + "/u2f-register?identity_token=" + token,
          jar: jar
        });
      })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        return requestAsync.getAsync({
          url: BASE_URL + "/2ndfactor/u2f/register_request",
          jar: jar,
        });
      })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        return requestAsync.postAsync({
          url: BASE_URL + "/2ndfactor/u2f/register",
          jar: jar,
          form: {
            s: "test"
          }
        });
      });
  }

  function execute_first_factor(jar: request.CookieJar) {
    return requestAsync.postAsync({
      url: BASE_URL + "/1stfactor",
      jar: jar,
      form: {
        username: "test_ok",
        password: "password"
      }
    });
  }

  function execute_failing_first_factor(jar: request.CookieJar) {
    return requestAsync.postAsync({
      url: BASE_URL + "/1stfactor",
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
    reset_password: execute_reset_password,
    u2f_authentication: execute_u2f_authentication,
    u2f_registration: execute_u2f_registration,
    first_factor: execute_first_factor,
    failing_first_factor: execute_failing_first_factor,
    totp: execute_totp,
    register_totp: execute_register_totp,
  };
};

