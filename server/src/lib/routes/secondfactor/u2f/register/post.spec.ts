import * as Express from "express";
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import U2FRegisterPost = require("./post");
import { AuthenticationSessionHandler } from "../../../../AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../../../../types/AuthenticationSession";
import ExpressMock = require("../../../../stubs/express.spec");
import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../../../ServerVariables";


describe("routes/secondfactor/u2f/register/post", function () {
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;
  let authSession: AuthenticationSession;

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

    authSession = AuthenticationSessionHandler.get(req as any, vars.logger);
  });

  describe("test registration", test_registration);


  function test_registration() {
    it("should save u2f meta and return status code 200", function () {
      const expectedStatus = {
        keyHandle: "keyHandle",
        publicKey: "pbk",
        certificate: "cert"
      };
      mocks.u2f.checkRegistrationStub.returns(BluebirdPromise.resolve(expectedStatus));

      authSession.register_request = {
        appId: "app",
        challenge: "challenge",
        keyHandle: "key",
        version: "U2F_V2"
      };
      return U2FRegisterPost.default(vars)(req as any, res as any)
        .then(function () {
          assert.equal("user", mocks.userDataStore.saveU2FRegistrationStub.getCall(0).args[0]);
          assert.equal(authSession.identity_check, undefined);
        });
    });

    it("should return error message on finishRegistration error", function () {
      mocks.u2f.checkRegistrationStub.returns({ errorCode: 500 });

      authSession.register_request = {
        appId: "app",
        challenge: "challenge",
        keyHandle: "key",
        version: "U2F_V2"
      };

      return U2FRegisterPost.default(vars)(req as any, res as any)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function () {
          assert.equal(200, res.status.getCall(0).args[0]);
          assert.deepEqual(res.send.getCall(0).args[0], {
            error: "Operation failed."
          });
          return BluebirdPromise.resolve();
        });
    });

    it("should return error message when register_request is not provided", function () {
      mocks.u2f.checkRegistrationStub.returns(BluebirdPromise.resolve());
      authSession.register_request = undefined;
      return U2FRegisterPost.default(vars)(req as any, res as any)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function () {
          assert.equal(200, res.status.getCall(0).args[0]);
          assert.deepEqual(res.send.getCall(0).args[0], {
            error: "Operation failed."
          });
          return BluebirdPromise.resolve();
        });
    });

    it("should return error message when no auth request has been initiated", function () {
      mocks.u2f.checkRegistrationStub.returns(BluebirdPromise.resolve());
      authSession.register_request = undefined;
      return U2FRegisterPost.default(vars)(req as any, res as any)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function () {
          assert.equal(200, res.status.getCall(0).args[0]);
          assert.deepEqual(res.send.getCall(0).args[0], {
            error: "Operation failed."
          });
          return BluebirdPromise.resolve();
        });
    });

    it("should return error message when identity has not been verified", function () {
      authSession.identity_check = undefined;
      return U2FRegisterPost.default(vars)(req as any, res as any)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function () {
          assert.equal(200, res.status.getCall(0).args[0]);
          assert.deepEqual(res.send.getCall(0).args[0], {
            error: "Operation failed."
          });
          return BluebirdPromise.resolve();
        });
    });
  }
});

