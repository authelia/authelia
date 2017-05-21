
import sinon = require("sinon");
import Promise = require("bluebird");
import assert = require("assert");
import express = require("express");

import ExpressMock = require("../mocks/express");
import DenyNotLogged = require("../../../src/lib/routes/DenyNotLogged");

describe("test not logged", function () {
  it("should return status code 403 when auth_session has not been previously created", function () {
    return test_auth_session_not_created();
  });

  it("should return status code 403 when auth_session has failed first factor", function () {
    return test_auth_first_factor_not_validated();
  });

  it("should return status code 204 when auth_session has succeeded first factor stage", function () {
    return test_auth_with_first_factor_validated();
  });
});

function test_auth_session_not_created() {
  return new Promise(function (resolve, reject) {
    const send = sinon.spy(resolve);
    const status = sinon.spy(function (code: number) {
      assert.equal(403, code);
    });
    const req = ExpressMock.RequestMock();
    const res = ExpressMock.ResponseMock();
    req.session = {};
    res.send = send;
    res.status = status;

    DenyNotLogged(reject)(req as any, res as any);
  });
}

function test_auth_first_factor_not_validated() {
  return new Promise(function (resolve, reject) {
    const send = sinon.spy(resolve);
    const status = sinon.spy(function (code: number) {
      assert.equal(403, code);
    });
    const req = {
      session: {
        auth_session: {
          first_factor: false,
          second_factor: false
        }
      }
    };

    const res = {
      send: send,
      status: status
    };

    DenyNotLogged(reject)(req as any, res as any);
  });
}

function test_auth_with_first_factor_validated() {
  return new Promise(function (resolve, reject) {
    const req = {
      session: {
        auth_session: {
          first_factor: true,
          second_factor: false
        }
      }
    };

    const res = {
      send: sinon.spy(),
      status: sinon.spy()
    };

    DenyNotLogged(resolve)(req as any, res as any);
  });
}
