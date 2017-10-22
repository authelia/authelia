
import BluebirdPromise = require("bluebird");
import Sinon = require("sinon");
import Assert = require("assert");
import Exceptions = require("../../../../../src/lib/Exceptions");
import { AuthenticationSessionHandler } from "../../../../../src/lib/AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../../../types/AuthenticationSession";
import SignPost = require("../../../../../src/lib/routes/secondfactor/totp/sign/post");
import { ServerVariables } from "../../../../../src/lib/ServerVariables";

import ExpressMock = require("../../../../mocks/express");
import { UserDataStoreStub } from "../../../../mocks/storage/UserDataStoreStub";
import { ServerVariablesMock, ServerVariablesMockBuilder } from "../../../../mocks/ServerVariablesMockBuilder";

describe("test totp route", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let authSession: AuthenticationSession;
  let vars: ServerVariables;
  let mocks: ServerVariablesMock;

  beforeEach(function () {
    const s = ServerVariablesMockBuilder.build();
    vars = s.variables;
    mocks = s.mocks;
    const app_get = Sinon.stub();
    req = {
      app: {},
      body: {
        token: "abc"
      },
      session: {},
      query: {
        redirect: "http://redirect"
      }
    };
    res = ExpressMock.ResponseMock();

    const doc = {
      userid: "user",
      secret: {
        base32: "ABCDEF"
      }
    };
    mocks.userDataStore.retrieveTOTPSecretStub.returns(BluebirdPromise.resolve(doc));
    authSession = AuthenticationSessionHandler.get(req as any, vars.logger);
    authSession.userid = "user";
    authSession.first_factor = true;
    authSession.second_factor = false;
  });


  it("should send status code 200 when totp is valid", function () {
    mocks.totpHandler.validateStub.returns(true);
    return SignPost.default(vars)(req as any, res as any)
      .then(function () {
        Assert.equal(true, authSession.second_factor);
        return BluebirdPromise.resolve();
      });
  });

  it("should send error message when totp is not valid", function () {
    mocks.totpHandler.validateStub.returns(false);
    return SignPost.default(vars)(req as any, res as any)
      .then(function () {
        Assert.equal(false, authSession.second_factor);
        Assert.equal(res.status.getCall(0).args[0], 200);
        Assert.deepEqual(res.send.getCall(0).args[0], {
          error: "Operation failed."
        });
        return BluebirdPromise.resolve();
      });
  });
});

