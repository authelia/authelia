
import BluebirdPromise = require("bluebird");
import Sinon = require("sinon");
import assert = require("assert");
import winston = require("winston");

import exceptions = require("../../../../../src/lib/Exceptions");
import AuthenticationSessionHandler = require("../../../../../src/lib/AuthenticationSession");
import { AuthenticationSession } from "../../../../../types/AuthenticationSession";
import SignPost = require("../../../../../src/lib/routes/secondfactor/totp/sign/post");
import { ServerVariables } from "../../../../../src/lib/ServerVariables";

import ExpressMock = require("../../../../mocks/express");
import { UserDataStoreStub } from "../../../../mocks/storage/UserDataStoreStub";
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../../../mocks/ServerVariablesMockBuilder";

describe("test totp route", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let authSession: AuthenticationSession;
  let vars: ServerVariables;
  let mocks: ServerVariablesMock;

  beforeEach(function () {
    const s = ServerVariablesMockBuilder.build();
    vars = s.variables;
    mocks = s.mocks;
    const app_get = Sinon.stub();
    req = {
      app: {
        get: Sinon.stub().returns({ logger: winston })
      },
      body: {
        token: "abc"
      },
      session: {},
      query: {
        redirect: "http://redirect"
      }
    };
    res = ExpressMock.ResponseMock();

    const doc = {
      userid: "user",
      secret: {
        base32: "ABCDEF"
      }
    };
    mocks.userDataStore.retrieveTOTPSecretStub.returns(BluebirdPromise.resolve(doc));
    return AuthenticationSessionHandler.get(req as any, vars.logger)
      .then(function (_authSession) {
        authSession = _authSession;
        authSession.userid = "user";
        authSession.first_factor = true;
        authSession.second_factor = false;
      });
  });


  it("should send status code 200 when totp is valid", function () {
    mocks.totpHandler.validateStub.returns(true);
    return SignPost.default(vars)(req as any, res as any)
      .then(function () {
        assert.equal(true, authSession.second_factor);
        return BluebirdPromise.resolve();
      });
  });

  it("should send error message when totp is not valid", function () {
    mocks.totpHandler.validateStub.returns(false);
    return SignPost.default(vars)(req as any, res as any)
      .then(function () {
        assert.equal(false, authSession.second_factor);
        assert.equal(res.status.getCall(0).args[0], 200);
        assert.deepEqual(res.send.getCall(0).args[0], {
          error: "Operation failed."
        });
        return BluebirdPromise.resolve();
      });
  });

  it("should send status code 401 when session has not been initiated", function () {
    mocks.totpHandler.validateStub.returns(true);
    req.session = {};
    return SignPost.default(vars)(req as any, res as any)
      .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
      .catch(function () {
        assert.equal(401, res.status.getCall(0).args[0]);
        return BluebirdPromise.resolve();
      });
  });
});

