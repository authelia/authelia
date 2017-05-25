
import BluebirdPromise = require("bluebird");
import request = require("request");
import assert = require("assert");
import express = require("express");
import nodemailer = require("nodemailer");
import Endpoints = require("../../src/server/endpoints");

import NodemailerMock = require("./mocks/nodemailer");

const requestAsync: typeof request = BluebirdPromise.promisifyAll(request) as typeof request;

export = function (port: number) {
  const PORT = port;
  const BASE_URL = "http://localhost:" + PORT;

  function execute_reset_password(jar: request.CookieJar, transporter: NodemailerMock.NodemailerTransporterMock, user: string, new_password: string) {
    return requestAsync.getAsync({
      url: BASE_URL + Endpoints.RESET_PASSWORD_IDENTITY_START_GET,
      jar: jar,
      qs: { userid: user }
    })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        const html_content = transporter.sendMail.getCall(0).args[0].html;
        const regexp = /identity_token=([a-zA-Z0-9]+)/;
        const token = regexp.exec(html_content)[1];
        // console.log(html_content, token);
        return requestAsync.getAsync({
          url: BASE_URL + Endpoints.RESET_PASSWORD_IDENTITY_FINISH_GET + "?identity_token=" + token,
          jar: jar
        });
      })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        return requestAsync.postAsync({
          url: BASE_URL + Endpoints.RESET_PASSWORD_FORM_POST,
          jar: jar,
          form: {
            password: new_password
          }
        });
      });
  }

  function execute_register_totp(jar: request.CookieJar, transporter: NodemailerMock.NodemailerTransporterMock) {
    return requestAsync.getAsync({
      url: BASE_URL + Endpoints.SECOND_FACTOR_TOTP_IDENTITY_START_GET,
      jar: jar
    })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        const html_content = transporter.sendMail.getCall(0).args[0].html;
        const regexp = /identity_token=([a-zA-Z0-9]+)/;
        const token = regexp.exec(html_content)[1];
        return requestAsync.getAsync({
          url: BASE_URL + Endpoints.SECOND_FACTOR_TOTP_IDENTITY_FINISH_GET + "?identity_token=" + token,
          jar: jar
        });
      })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        const regex = /<p id="secret">([A-Z0-9]+)<\/p>/g;
        const secret = regex.exec(res.body);
        return BluebirdPromise.resolve(secret[1]);
      });
  }

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

  function execute_u2f_registration(jar: request.CookieJar, transporter: NodemailerMock.NodemailerTransporterMock) {
    return requestAsync.getAsync({
      url: BASE_URL + Endpoints.SECOND_FACTOR_U2F_IDENTITY_START_GET,
      jar: jar
    })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        const html_content = transporter.sendMail.getCall(0).args[0].html;
        const regexp = /identity_token=([a-zA-Z0-9]+)/;
        const token = regexp.exec(html_content)[1];
        // console.log(html_content, token);
        return requestAsync.getAsync({
          url: BASE_URL + Endpoints.SECOND_FACTOR_U2F_IDENTITY_FINISH_GET + "?identity_token=" + token,
          jar: jar
        });
      })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        return requestAsync.getAsync({
          url: BASE_URL + Endpoints.SECOND_FACTOR_U2F_REGISTER_REQUEST_GET,
          jar: jar,
        });
      })
      .then(function (res: request.RequestResponse) {
        assert.equal(res.statusCode, 200);
        return requestAsync.postAsync({
          url: BASE_URL + Endpoints.SECOND_FACTOR_U2F_REGISTER_POST,
          jar: jar,
          form: {
            s: "test"
          }
        });
      });
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
    reset_password: execute_reset_password,
    u2f_authentication: execute_u2f_authentication,
    u2f_registration: execute_u2f_registration,
    first_factor: execute_first_factor,
    failing_first_factor: execute_failing_first_factor,
    totp: execute_totp,
    register_totp: execute_register_totp,
  };
};

