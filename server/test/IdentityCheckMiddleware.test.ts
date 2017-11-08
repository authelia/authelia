
import sinon = require("sinon");
import IdentityValidator = require("../src/lib/IdentityCheckMiddleware");
import { AuthenticationSessionHandler }
  from "../src/lib/AuthenticationSessionHandler";
import { AuthenticationSession } from "../types/AuthenticationSession";
import { UserDataStore } from "../src/lib/storage/UserDataStore";
import exceptions = require("../src/lib/Exceptions");
import { ServerVariables } from "../src/lib/ServerVariables";
import Assert = require("assert");
import express = require("express");
import BluebirdPromise = require("bluebird");
import ExpressMock = require("./mocks/express");
import NotifierMock = require("./mocks/Notifier");
import IdentityValidatorMock = require("./mocks/IdentityValidator");
import { RequestLoggerStub } from "./mocks/RequestLoggerStub";
import { ServerVariablesMock, ServerVariablesMockBuilder }
  from "./mocks/ServerVariablesMockBuilder";
import { PRE_VALIDATION_TEMPLATE }
  from "../src/lib/IdentityCheckPreValidationTemplate";


describe("test identity check process", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let app: express.Application;
  let app_get: sinon.SinonStub;
  let app_post: sinon.SinonStub;
  let identityValidable: IdentityValidatorMock.IdentityValidableMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();

    identityValidable = IdentityValidatorMock.IdentityValidableMock();

    req.headers = {};
    req.originalUrl = "/non-api/xxx";
    req.session = {};

    req.query = {};
    req.app = {};

    mocks.notifier.notifyStub.returns(BluebirdPromise.resolve());
    mocks.userDataStore.produceIdentityValidationTokenStub
      .returns(Promise.resolve());
    mocks.userDataStore.consumeIdentityValidationTokenStub
      .returns(Promise.resolve({ userId: "user" }));

    app = express();
    app_get = sinon.stub(app, "get");
    app_post = sinon.stub(app, "post");
  });

  afterEach(function () {
    app_get.restore();
    app_post.restore();
  });

  describe("test start GET", function () {
    it("should redirect to error 401 if pre validation initialization \
throws a first factor error", function () {
        identityValidable.preValidationInit.returns(BluebirdPromise.reject(
          new exceptions.FirstFactorValidationError(
            "Error during prevalidation")));
        const callback = IdentityValidator.get_start_validation(
          identityValidable, "/endpoint", vars);

        return callback(req as any, res as any, undefined)
          .then(function () { return BluebirdPromise.reject("Should fail"); })
          .catch(function () {
            Assert(res.redirect.calledWith("/error/401"));
          });
      });

    // In that case we answer with 200 to avoid user enumeration.
    it("should send 200 if email is missing in provided identity", function () {
      const identity = { userid: "abc" };

      identityValidable.preValidationInit
        .returns(BluebirdPromise.resolve(identity));
      const callback = IdentityValidator
        .get_start_validation(identityValidable, "/endpoint", vars);

      return callback(req as any, res as any, undefined)
        .then(function () {
          Assert(identityValidable.preValidationResponse.called);
        });
    });

    // In that case we answer with 200 to avoid user enumeration.
    it("should send 200 if userid is missing in provided identity",
      function () {
        const endpoint = "/protected";
        const identity = { email: "abc@example.com" };

        identityValidable.preValidationInit
          .returns(BluebirdPromise.resolve(identity));
        const callback = IdentityValidator
          .get_start_validation(identityValidable, "/endpoint", vars);

        return callback(req as any, res as any, undefined)
          .then(function () {
            Assert(identityValidable.preValidationResponse.called);
          });
      });

    it("should issue a token, send an email and return 204", function () {
      const endpoint = "/protected";
      const identity = { userid: "user", email: "abc@example.com" };
      req.get = sinon.stub().withArgs("Host").returns("localhost");

      identityValidable.preValidationInit
        .returns(BluebirdPromise.resolve(identity));
      const callback = IdentityValidator
        .get_start_validation(identityValidable, "/finish_endpoint", vars);

      return callback(req as any, res as any, undefined)
        .then(function () {
          Assert(mocks.notifier.notifyStub.calledOnce);
          Assert(mocks.userDataStore.produceIdentityValidationTokenStub
            .calledOnce);
          Assert.equal(mocks.userDataStore.produceIdentityValidationTokenStub
            .getCall(0).args[0], "user");
          Assert.equal(mocks.userDataStore.produceIdentityValidationTokenStub
            .getCall(0).args[3], 240000);
        });
    });
  });



  describe("test finish GET", function () {
    it("should send 401 if no identity_token is provided", function () {

      const callback = IdentityValidator
        .get_finish_validation(identityValidable, vars);

      return callback(req as any, res as any, undefined)
        .then(function () {
          return BluebirdPromise.reject("Should fail");
        })
        .catch(function () {
          Assert(res.redirect.calledWith("/error/401"));
        });
    });

    it("should call postValidation if identity_token is provided and still \
valid", function () {
        req.query.identity_token = "token";

        const callback = IdentityValidator
          .get_finish_validation(identityValidable, vars);
        return callback(req as any, res as any, undefined);
      });

    it("should return 401 if identity_token is provided but invalid",
      function () {
        req.query.identity_token = "token";

        mocks.userDataStore.consumeIdentityValidationTokenStub
          .returns(BluebirdPromise.reject(new Error("Invalid token")));

        const callback = IdentityValidator
          .get_finish_validation(identityValidable, vars);
        return callback(req as any, res as any, undefined)
          .then(function () {
            return BluebirdPromise.reject("Should fail");
          })
          .catch(function () {
            Assert(res.redirect.calledWith("/error/401"));
          });
      });

    it("should set the identity_check session object even if session does \
not exist yet", function () {
        req.query.identity_token = "token";

        req.session = {};
        const authSession =
          AuthenticationSessionHandler.get(req as any, vars.logger);
        const callback = IdentityValidator
          .get_finish_validation(identityValidable, vars);

        return callback(req as any, res as any, undefined)
          .then(function () {
            return BluebirdPromise.reject("Should fail");
          })
          .catch(function () {
            Assert.equal(authSession.identity_check.userid, "user");
          });
      });
  });
});
