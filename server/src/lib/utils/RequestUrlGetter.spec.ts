import RequestUrlGetter from "./RequestUrlGetter";
import * as Assert from "assert";
import * as Sinon from "sinon";
import { RequestLoggerStub } from "../logging/RequestLoggerStub.spec";

describe('RequestUrlGetter', function() {
  let req: any;
  beforeEach(function() {
    req = {
      app: {
        get: Sinon.stub().returns({
          logger: new RequestLoggerStub()
        })
      },
      headers: {}
    }
  })

  it("should return the content of X-Original-Uri header", function() {
    req.headers["x-original-url"] = "https://mytarget.example.com";
    Assert.equal(RequestUrlGetter.getOriginalUrl(req), "https://mytarget.example.com");
  })

  describe("Use X-Forwarded-Proto, X-Forwarded-Host and X-Forwarded-Uri headers", function() {
    it("should get URL from Forwarded headers", function() {
      req.headers["x-forwarded-proto"] = "https";
      req.headers["x-forwarded-host"] = "mytarget.example.com";
      req.headers["x-forwarded-uri"] = "/"
      Assert.equal(RequestUrlGetter.getOriginalUrl(req), "https://mytarget.example.com/");
    })

    it("should get URL from Forwarded headers without URI", function() {
      req.headers["x-forwarded-proto"] = "https";
      req.headers["x-forwarded-host"] = "mytarget.example.com";
      Assert.equal(RequestUrlGetter.getOriginalUrl(req), "https://mytarget.example.com");
    })
  });

  it("should throw when no header is provided", function() {
    Assert.throws(() => {
      RequestUrlGetter.getOriginalUrl(req)
    })
  })

  it("should throw when only some of X-Forwarded-* headers are provided", function() {
    req.headers["x-forwarded-proto"] = "https";
    Assert.throws(() => {
      RequestUrlGetter.getOriginalUrl(req)
    })
  })
})