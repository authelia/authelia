import * as Express from "express";
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");
import U2FRegisterRequestGet = require("./get");
import ExpressMock = require("../../../../stubs/express.spec");
import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../../../ServerVariables";

describe("routes/secondfactor/u2f/register_request/get", function () {
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

    mocks.userDataStore.saveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));
    mocks.userDataStore.retrieveU2FRegistrationStub.returns(BluebirdPromise.resolve({}));

    res = ExpressMock.ResponseMock();
    res.send = sinon.spy();
    res.json = sinon.spy();
    res.status = sinon.spy();
  });

  describe("test registration request", () => {
    it("should send back the registration request and save it in the session", function () {
      const expectedRequest = {
        test: "abc"
      };
      mocks.u2f.requestStub.returns(BluebirdPromise.resolve(expectedRequest));
      return U2FRegisterRequestGet.default(vars)(req as any, res as any)
        .then(function () {
          Assert.deepEqual(expectedRequest, res.json.getCall(0).args[0]);
        });
    });

    it("should return internal error on registration request", function () {
      res.send = sinon.spy();
      mocks.u2f.requestStub.returns(BluebirdPromise.reject("Internal error"));
      return U2FRegisterRequestGet.default(vars)(req as any, res as any)
        .then(function () {
          Assert.equal(res.status.getCall(0).args[0], 200);
          Assert.deepEqual(res.send.getCall(0).args[0], {
            error: "Operation failed."
          });
        });
    });

    it("should return forbidden if identity has not been verified", function () {
      req.session.auth.identity_check = undefined;
      return U2FRegisterRequestGet.default(vars)(req as any, res as any)
        .then(function () {
          Assert.equal(200, res.status.getCall(0).args[0]);
          Assert.deepEqual(res.send.getCall(0).args[0], {
            error: "Operation failed."
          });
        });
    });
  });
});

