import * as Express from "express";
import * as Bluebird from "bluebird";
import { ServerVariables } from "../../../ServerVariables";
import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../../ServerVariablesMockBuilder.spec";
import * as ExpressMock from "../../../stubs/express.spec";
import Post from "./Post";
import * as Assert from "assert";

describe("routes/secondfactor/preferences/Post", function() {
  let vars: ServerVariables;
  let mocks: ServerVariablesMock;
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;

  beforeEach(function() {
    const sv = ServerVariablesMockBuilder.build();
    vars = sv.variables;
    mocks = sv.mocks;

    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
  })

  it("should save the method in DB", async function() {
    mocks.userDataStore.savePrefered2FAMethodStub.returns(Bluebird.resolve());
    req.body.method = 'totp';
    req.session.auth = {
      userid: 'john'
    }
    await Post(vars)(req, res as any);
    Assert(mocks.userDataStore.savePrefered2FAMethodStub.calledWith('john', 'totp'));
    Assert(res.status.calledWith(204));
    Assert(res.send.calledWith());
  });

  it("should fail if no method is provided in body", async function() {
    req.session.auth = {
      userid: 'john'
    }
    await Post(vars)(req, res as any);
    Assert(res.status.calledWith(200));
    Assert(res.send.calledWith({ error: "Operation failed." }));
  });

  it("should fail if access to DB fails", async function() {
    mocks.userDataStore.savePrefered2FAMethodStub.returns(Bluebird.reject(new Error('DB access failed.')));
    req.body.method = 'totp'
    req.session.auth = {
      userid: 'john'
    }
    await Post(vars)(req, res as any);
    Assert(res.status.calledWith(200));
    Assert(res.send.calledWith({ error: "Operation failed." }));
  });
});