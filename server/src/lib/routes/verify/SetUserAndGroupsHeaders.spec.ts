import * as ExpressMock from "../../stubs/express.spec";
import * as Assert from "assert";
import { HEADER_REMOTE_USER, HEADER_REMOTE_GROUPS } from "../../constants";
import SetUserAndGroupsHeaders from "./SetUserAndGroupsHeaders";

describe("routes/verify/SetUserAndGroupsHeaders", function() {
  it('should set the correct headers', function() {
    const res = ExpressMock.ResponseMock();
    SetUserAndGroupsHeaders(res as any, "john", ["group1", "group2"]);
    Assert(res.setHeader.calledWith(HEADER_REMOTE_USER, "john"));
    Assert(res.setHeader.calledWith(HEADER_REMOTE_GROUPS, "group1,group2"));
  })
})