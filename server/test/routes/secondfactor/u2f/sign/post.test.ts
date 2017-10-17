
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");
import U2FSignPost = require("../../../../../src/lib/routes/secondfactor/u2f/sign/post");
import AuthenticationSession = require("../../../../../src/lib/AuthenticationSession");
import { ServerVariables } from "../../../../../src/lib/ServerVariables";
import winston = require("winston");

import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../../../mocks/ServerVariablesMockBuilder";
import ExpressMock = require("../../../../mocks/express");
import U2FMock = require("../../../../mocks/u2f");
import U2f = require("u2f");

describe("test u2f routes: sign", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.app = {};

    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    req.session = {
      auth: {
        userid: "user",
        first_factor: true,
        second_factor: false,
        identity_check: {
          challenge: "u2f-register",
          userid: "user"
        }
      }
    };
    req.headers = {};
    req.headers.host = "localhost";

    const options = {
      inMemoryOnly: true
    };

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  });

  it("should return status code 204", function () {
    const expectedStatus = {
      keyHandle: "keyHandle",
      publicKey: "pbk",
      certificate: "cert"
    };
    mocks.u2f.checkSignatureStub.returns(expectedStatus);

    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({
      registration: {
        publicKey: "PUBKEY"
      }
    }));

    req.session.auth.sign_request = {
      appId: "app",
      challenge: "challenge",
      keyHandle: "key",
      version: "U2F_V2"
    };
    return U2FSignPost.default(vars)(req as any, res as any)
      .then(function () {
        Assert(req.session.auth.second_factor);
      });
  });

  it("should return unauthorized error on registration request internal error", function () {
    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({
      registration: {
        publicKey: "PUBKEY"
      }
    }));
    mocks.u2f.checkSignatureStub.returns({ errorCode: 500 });

    req.session.auth.sign_request = {
      appId: "app",
      challenge: "challenge",
      keyHandle: "key",
      version: "U2F_V2"
    };
    return U2FSignPost.default(vars)(req as any, res as any)
      .then(function () {
        Assert.equal(res.status.getCall(0).args[0], 200);
        Assert.deepEqual(res.send.getCall(0).args[0],
          { error: "Operation failed." });
      });
  });
});

