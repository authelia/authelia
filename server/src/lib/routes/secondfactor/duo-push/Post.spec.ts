import * as Express from "express";
import { ServerVariables } from "../../../ServerVariables";
import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../../ServerVariablesMockBuilder.spec";
import * as ExpressMock from "../../../stubs/express.spec";
import Post from "./Post";
import * as Sinon from "sinon";
import * as Assert from "assert";
import { Level } from "../../../authentication/Level";
const DuoApi = require("@duosecurity/duo_api");


describe("routes/secondfactor/duo-push/Post", function() {
  let vars: ServerVariables;
  let mocks: ServerVariablesMock;
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;

  beforeEach(function() {
    const sv = ServerVariablesMockBuilder.build();
    vars = sv.variables;
    mocks = sv.mocks;

    vars.config.duo_api = {
      hostname: 'abc',
      integration_key: 'xyz',
      secret_key: 'secret',
    };

    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
  });

  it("should raise authentication level of user", async function() {
    const mock = Sinon.stub(DuoApi, "Client");
    mock.returns({
      jsonApiCall: Sinon.stub().yields({response: {result: 'allow'}})
    });
    req.session.auth = {
      userid: 'john'
    };

    Assert.equal(req.session.auth.authentication_level, undefined);
    await Post(vars)(req, res as any);
    Assert(res.status.calledWith(204));
    Assert(res.send.calledWith());
    Assert.equal(req.session.auth.authentication_level, Level.TWO_FACTOR);
    mock.restore();
  });

  it("should block if no duo API is configured", async function() {
    const mock = Sinon.stub(DuoApi, "Client");
    mock.returns({
      jsonApiCall: Sinon.stub().yields({response: {result: 'allow'}})
    });
    req.session.auth = {
      userid: 'john'
    };
    vars.config.duo_api = undefined;

    Assert.equal(req.session.auth.authentication_level, undefined);
    await Post(vars)(req, res as any);
    Assert(res.status.calledWith(200));
    Assert(res.send.calledWith({error: 'Operation failed.'}));
    Assert.equal(req.session.auth.authentication_level, undefined);
    mock.restore();
  });

  it("should block if user denied notification", async function() {
    const mock = Sinon.stub(DuoApi, "Client");
    mock.returns({
      jsonApiCall: Sinon.stub().yields({response: {result: 'deny'}})
    });
    req.session.auth = {
      userid: 'john'
    };

    Assert.equal(req.session.auth.authentication_level, undefined);
    await Post(vars)(req, res as any);
    Assert(res.status.calledWith(200));
    Assert(res.send.calledWith({error: 'Operation failed.'}));
    Assert.equal(req.session.auth.authentication_level, undefined);
    mock.restore();
  });

  it("should block if duo push service is down", function() {
    const mock = Sinon.stub(DuoApi, "Client");
    const timerMock = Sinon.useFakeTimers();
    mock.returns({
      jsonApiCall: Sinon.stub()
    });
    req.session.auth = {
      userid: 'john'
    };

    Assert.equal(req.session.auth.authentication_level, undefined);
    const promise = Post(vars)(req, res as any)
      .then(() => {
        Assert(res.status.calledWith(200));
        Assert(res.send.calledWith({error: 'Operation failed.'}));
        Assert.equal(req.session.auth.authentication_level, undefined);

        mock.restore();
        timerMock.restore();
      });
    // Move forward in time to timeout.
    timerMock.tick(62000);
    return promise;
  });
});