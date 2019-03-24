import * as Express from "express";
import * as Bluebird from "bluebird";
import { ServerVariables } from "../../../ServerVariables";
import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../../ServerVariablesMockBuilder.spec";
import * as ExpressMock from "../../../stubs/express.spec";
import Get from "./Get";
import * as Assert from "assert";

describe("routes/secondfactor/preferences/Get", function() {
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

  it("should get the method from db", async function() {
    mocks.userDataStore.retrievePrefered2FAMethodStub.returns(Bluebird.resolve('totp'));
    await Get(vars)(req, res as any);
    Assert(res.json.calledWith({method: 'totp'}));
  });

  it("should fail when database fail to retrieve method", async function() {
    mocks.userDataStore.retrievePrefered2FAMethodStub.returns(Bluebird.reject(new Error('DB connection failed.')));
    await Get(vars)(req, res as any);
    Assert(res.status.calledWith(200));
    Assert(res.send.calledWith({ error: "Operation failed." }));
  })
});