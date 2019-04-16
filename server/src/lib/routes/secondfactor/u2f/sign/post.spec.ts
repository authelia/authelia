import * as Express from "express";
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");
import U2FSignPost = require("./post");
import { ServerVariables } from "../../../../ServerVariables";
import UserMessages = require("../../../../UserMessages");
import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../../../ServerVariablesMockBuilder.spec";
import ExpressMock = require("../../../../stubs/express.spec");
import { Level } from "../../../../authentication/Level";

describe("routes/secondfactor/u2f/sign/post", function () {
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.originalUrl = "/api/xxxx";

    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    req.session = {
      ...req.session,
      auth: {
        userid: "user",
        authentication_level: Level.ONE_FACTOR,
        identity_check: {
          challenge: "u2f-register",
          userid: "user"
        }
      }
    };
    req.headers = {};
    req.headers.host = "localhost";

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
        Assert.equal(req.session.auth.authentication_level, Level.TWO_FACTOR);
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
          { error: UserMessages.OPERATION_FAILED });
      });
  });
});

