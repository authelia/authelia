import * as Express from "express";
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import U2FSignRequestGet = require("./get");
import ExpressMock = require("../../../../stubs/express.spec");
import { Request } from "u2f";
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../../../ServerVariables";

describe("routes/secondfactor/u2f/sign_request/get", function () {
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.originalUrl = "/api/xxxx";
    req.session = {
      ...req.session,
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

    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  });

  it("should send back the sign request and save it in the session", function () {
    const expectedRequest: Request = {
      version: "U2F_V2",
      appId: 'app',
      challenge: 'challenge!'
    };
    mocks.u2f.requestStub.returns(expectedRequest);
    mocks.userDataStore.retrieveU2FRegistrationStub
      .returns(BluebirdPromise.resolve({
        registration: {
          keyHandle: "KeyHandle"
        }
      }));

    return U2FSignRequestGet.default(vars)(req as any, res as any)
      .then(() => {
        assert.deepEqual(expectedRequest, req.session.auth.sign_request);
        assert.deepEqual(expectedRequest, res.json.getCall(0).args[0]);
      });
  });
});

