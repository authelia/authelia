import * as Express from "express";
import GetHeader from "./GetHeader";
import { RequestMock } from "../stubs/express.spec";
import * as Assert from "assert";

describe('GetHeader', function() {
  let req: Express.Request;
  beforeEach(() => {
    req = RequestMock();
  });

  it('should return the header if it exists', function() {
    req.headers["x-target-url"] = 'www.example.com';
    Assert.equal(GetHeader(req, 'x-target-url'), 'www.example.com');
  });

  it('should return undefined if header does not exist', function() {
    Assert.equal(GetHeader(req, 'x-target-url'), undefined);
  });
});