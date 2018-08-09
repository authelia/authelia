
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import U2FSignRequestGet = require("./get");
import ExpressMock = require("../../../../stubs/express.spec");
import { UserDataStoreStub } from "../../../../storage/UserDataStoreStub.spec";
import U2FMock = require("../../../../stubs/u2f.spec");
import U2f = require("u2f");
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../../../ServerVariables";

import { SignMessage } from "../../../../../../../shared/SignMessage";

describe("routes/secondfactor/u2f/sign_request/get", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    req.originalUrl = "/api/xxxx";
    req.app = {};
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

    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    const options = {
      inMemoryOnly: true
    };

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  });

  it("should send back the sign request and save it in the session", function () {
    const expectedRequest: U2f.RegistrationResult = {
      keyHandle: "keyHandle",
      publicKey: "publicKey",
      certificate: "Certificate",
      successful: true
    };
    mocks.u2f.requestStub.returns(expectedRequest);
    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({
      registration: {
        publicKey: "PUBKEY",
        keyHandle: "KeyHandle"
      }
    }));

    return U2FSignRequestGet.default(vars)(req as any, res as any)
      .then(function () {
        assert.deepEqual(expectedRequest, req.session.auth.sign_request);
        assert.deepEqual(expectedRequest, res.json.getCall(0).args[0].request);
      });
  });
});

