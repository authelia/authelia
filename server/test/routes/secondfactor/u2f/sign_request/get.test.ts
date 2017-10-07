
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import U2FSignRequestGet = require("../../../../../src/lib/routes/secondfactor/u2f/sign_request/get");
import AuthenticationSession = require("../../../../../src/lib/AuthenticationSession");
import { ServerVariablesHandler } from "../../../../../src/lib/ServerVariablesHandler";
import winston = require("winston");

import ExpressMock = require("../../../../mocks/express");
import { UserDataStoreStub } from "../../../../mocks/storage/UserDataStoreStub";
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");
import U2FMock = require("../../../../mocks/u2f");
import U2f = require("u2f");

import { SignMessage } from "../../../../../../shared/SignMessage";

describe("test u2f routes: sign_request", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock.ServerVariablesMock;
  let authSession: AuthenticationSession.AuthenticationSession;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.app = {};

    mocks = ServerVariablesMock.mock(req.app);

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

  it("should send back the sign request and save it in the session", function () {
    const expectedRequest: U2f.RegistrationResult = {
      keyHandle: "keyHandle",
      publicKey: "publicKey",
      certificate: "Certificate",
      successful: true
    };
    const u2f_mock = U2FMock.U2FMock();
    u2f_mock.request.returns(expectedRequest);

    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({
      registration: {
        publicKey: "PUBKEY",
        keyHandle: "KeyHandle"
      }
    }));

    mocks.u2f = u2f_mock;
    return U2FSignRequestGet.default(req as any, res as any)
      .then(function () {
        assert.deepEqual(expectedRequest, authSession.sign_request);
        assert.deepEqual(expectedRequest, res.json.getCall(0).args[0].request);
      });
  });
});

