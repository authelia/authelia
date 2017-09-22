
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");
import U2FSignPost = require("../../../../../../../src/server/lib/routes/secondfactor/u2f/sign/post");
import AuthenticationSession = require("../../../../../../../src/server/lib/AuthenticationSession");
import { ServerVariablesHandler } from "../../../../../../../src/server/lib/ServerVariablesHandler";
import winston = require("winston");

import ExpressMock = require("../../../../mocks/express");
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");
import U2FMock = require("../../../../mocks/u2f");
import U2f = require("u2f");

describe("test u2f routes: sign", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let authSession: AuthenticationSession.AuthenticationSession;
  let mocks: ServerVariablesMock.ServerVariablesMock;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.app = {};

    mocks = ServerVariablesMock.mock(req.app);
    mocks.logger = winston;

    req.session = {};
    AuthenticationSession.reset(req as any);
    req.headers = {};
    req.headers.host = "localhost";

    const options = {
      inMemoryOnly: true
    };

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();

    return AuthenticationSession.get(req as any)
      .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
        authSession = _authSession;
        authSession.userid = "user";
        authSession.first_factor = true;
        authSession.second_factor = false;
        authSession.identity_check = {
          challenge: "u2f-register",
          userid: "user"
        };
      });
  });

  it("should return status code 204", function () {
    const expectedStatus = {
      keyHandle: "keyHandle",
      publicKey: "pbk",
      certificate: "cert"
    };
    const u2f_mock = U2FMock.U2FMock();
    u2f_mock.checkSignature.returns(expectedStatus);

    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({
      registration: {
        publicKey: "PUBKEY"
      }
    }));

    authSession.sign_request = {
      appId: "app",
      challenge: "challenge",
      keyHandle: "key",
      version: "U2F_V2"
    };
    mocks.u2f = u2f_mock;
    return U2FSignPost.default(req as any, res as any)
      .then(function () {
        Assert(authSession.second_factor);
      });
  });

  it("should return unauthorized error on registration request internal error", function () {
    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({
      registration: {
        publicKey: "PUBKEY"
      }
    }));

    const u2f_mock = U2FMock.U2FMock();
    u2f_mock.checkSignature.returns({ errorCode: 500 });

    authSession.sign_request = {
      appId: "app",
      challenge: "challenge",
      keyHandle: "key",
      version: "U2F_V2"
    };
    mocks.u2f = u2f_mock;
    return U2FSignPost.default(req as any, res as any)
      .then(function () {
        Assert.equal(500, res.status.getCall(0).args[0]);
      });
  });
});

