
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import U2FSignRequestGet = require("../../../../../src/lib/routes/secondfactor/u2f/sign_request/get");
import AuthenticationSessionHandler = require("../../../../../src/lib/AuthenticationSession");
import { AuthenticationSession } from "../../../../../types/AuthenticationSession";
import ExpressMock = require("../../../../mocks/express");
import { UserDataStoreStub } from "../../../../mocks/storage/UserDataStoreStub";
import U2FMock = require("../../../../mocks/u2f");
import U2f = require("u2f");
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../../../mocks/ServerVariablesMockBuilder";
import { ServerVariables } from "../../../../../src/lib/ServerVariables";

import { SignMessage } from "../../../../../../shared/SignMessage";

describe("test u2f routes: sign_request", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
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

