import * as Express from "express";
import { ServerVariables } from "../../../ServerVariables";
import { ServerVariablesMockBuilder } from "../../../ServerVariablesMockBuilder.spec";
import * as ExpressMock from "../../../stubs/express.spec";
import Get from "./Get";
import * as Assert from "assert";


describe("routes/secondfactor/duo-push/Post", function() {
  let vars: ServerVariables;
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;

  beforeEach(function() {
    const sv = ServerVariablesMockBuilder.build();
    vars = sv.variables;

    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
  })

  it("should return default available methods", async function() {
    await Get(vars)(req, res as any);
    Assert(res.json.calledWith(["u2f", "totp"]));
  });

  it("should return duo as an available method", async function() {
    vars.config.duo_api = {
      hostname: "example.com",
      integration_key: "ABCDEFG",
      secret_key: "ekjfzelfjz",
    }
    await Get(vars)(req, res as any);
    Assert(res.json.calledWith(["u2f", "totp", "duo_push"]));
  });
});