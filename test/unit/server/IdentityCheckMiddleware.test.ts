
import sinon = require("sinon");
import IdentityValidator = require("../../../src/server/lib/IdentityCheckMiddleware");
import AuthenticationSession = require("../../../src/server/lib/AuthenticationSession");
import { UserDataStore } from "../../../src/server/lib/storage/UserDataStore";

import exceptions = require("../../../src/server/lib/Exceptions");
import assert = require("assert");
import Promise = require("bluebird");
import express = require("express");
import BluebirdPromise = require("bluebird");

import ExpressMock = require("./mocks/express");
import NotifierMock = require("./mocks/Notifier");
import IdentityValidatorMock = require("./mocks/IdentityValidator");
import ServerVariablesMock = require("./mocks/ServerVariablesMock");


describe("test identity check process", function () {
  let mocks: ServerVariablesMock.ServerVariablesMock;
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let notifier: NotifierMock.NotifierMock;
  let app: express.Application;
  let app_get: sinon.SinonStub;
  let app_post: sinon.SinonStub;
  let identityValidable: IdentityValidatorMock.IdentityValidableMock;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();

    identityValidable = IdentityValidatorMock.IdentityValidableMock();

    notifier = NotifierMock.NotifierMock();
    notifier.notify = sinon.stub().returns(Promise.resolve());

    req.headers = {};
    req.session = {};
    req.session = {};

    req.query = {};
    req.app = {};

    mocks = ServerVariablesMock.mock(req.app);
    mocks.notifier = notifier;

    mocks.userDataStore.produceIdentityValidationTokenStub.returns(Promise.resolve());
    mocks.userDataStore.consumeIdentityValidationTokenStub.returns(Promise.resolve({ userId: "user" }));

    app = express();
    app_get = sinon.stub(app, "get");
    app_post = sinon.stub(app, "post");
  });

  afterEach(function () {
    app_get.restore();
    app_post.restore();
  });

  describe("test start GET", test_start_get_handler);
  describe("test finish GET", test_finish_get_handler);

  function test_start_get_handler() {
    it("should send 401 if pre validation initialization throws a first factor error", function () {
      identityValidable.preValidationInit.returns(BluebirdPromise.reject(new exceptions.FirstFactorValidationError("Error during prevalidation")));
      const callback = IdentityValidator.get_start_validation(identityValidable, "/endpoint");

      return callback(req as any, res as any, undefined)
        .then(function () { return BluebirdPromise.reject("Should fail"); })
        .catch(function () {
          assert.equal(res.status.getCall(0).args[0], 401);
        });
    });

    it("should send 400 if email is missing in provided identity", function () {
      const identity = { userid: "abc" };

      identityValidable.preValidationInit.returns(BluebirdPromise.resolve(identity));
      const callback = IdentityValidator.get_start_validation(identityValidable, "/endpoint");

      return callback(req as any, res as any, undefined)
        .then(function () { return BluebirdPromise.reject("Should fail"); })
        .catch(function () {
          assert.equal(res.status.getCall(0).args[0], 400);
        });
    });

    it("should send 400 if userid is missing in provided identity", function () {
      const endpoint = "/protected";
      const identity = { email: "abc@example.com" };

      identityValidable.preValidationInit.returns(BluebirdPromise.resolve(identity));
      const callback = IdentityValidator.get_start_validation(identityValidable, "/endpoint");

      return callback(req as any, res as any, undefined)
        .then(function () { return BluebirdPromise.reject(new Error("It should fail")); })
        .catch(function (err: Error) {
          assert.equal(res.status.getCall(0).args[0], 400);
          return BluebirdPromise.resolve();
        });
    });

    it("should issue a token, send an email and return 204", function () {
      const endpoint = "/protected";
      const identity = { userid: "user", email: "abc@example.com" };
      req.get = sinon.stub().withArgs("Host").returns("localhost");

      identityValidable.preValidationInit.returns(BluebirdPromise.resolve(identity));
      const callback = IdentityValidator.get_start_validation(identityValidable, "/finish_endpoint");

      return callback(req as any, res as any, undefined)
        .then(function () {
          assert(notifier.notify.calledOnce);
          assert(mocks.userDataStore.produceIdentityValidationTokenStub.calledOnce);
          assert.equal(mocks.userDataStore.produceIdentityValidationTokenStub.getCall(0).args[0], "user");
          assert.equal(mocks.userDataStore.produceIdentityValidationTokenStub.getCall(0).args[3], 240000);
        });
    });
  }

  function test_finish_get_handler() {
    it("should send 403 if no identity_token is provided", function () {

      const callback = IdentityValidator.get_finish_validation(identityValidable);

      return callback(req as any, res as any, undefined)
        .then(function () { return BluebirdPromise.reject("Should fail"); })
        .catch(function () {
          assert.equal(res.status.getCall(0).args[0], 403);
        });
    });

    it("should call postValidation if identity_token is provided and still valid", function () {
      req.query.identity_token = "token";

      const callback = IdentityValidator.get_finish_validation(identityValidable);
      return callback(req as any, res as any, undefined);
    });

    it("should return 500 if identity_token is provided but invalid", function () {
      req.query.identity_token = "token";

      mocks.userDataStore.consumeIdentityValidationTokenStub.returns(BluebirdPromise.reject(new Error("Invalid token")));

      const callback = IdentityValidator.get_finish_validation(identityValidable);
      return callback(req as any, res as any, undefined)
        .then(function () { return BluebirdPromise.reject("Should fail"); })
        .catch(function () {
          assert.equal(res.status.getCall(0).args[0], 500);
        });
    });

    it("should set the identity_check session object even if session does not exist yet", function () {
      req.query.identity_token = "token";

      req.session = {};
      let authSession: AuthenticationSession.AuthenticationSession;
      const callback = IdentityValidator.get_finish_validation(identityValidable);

      return AuthenticationSession.get(req as any)
        .then(function (_authSession: AuthenticationSession.AuthenticationSession) {
          authSession = _authSession;
          return callback(req as any, res as any, undefined);
        })
        .then(function () { return BluebirdPromise.reject("Should fail"); })
        .catch(function () {
          assert.equal(authSession.identity_check.userid, "user");
          return BluebirdPromise.resolve();
        });
    });
  }
});
