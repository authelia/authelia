
import * as assert from "assert";
import * as Promise from "bluebird";
import * as sinon from "sinon";
import * as MockDate from "mockdate";
import UserDataStore from "../../../src/server/lib/UserDataStore";
import nedb = require("nedb");

describe("test user data store", function() {
  describe("test authentication traces", test_authentication_traces);
});

function test_authentication_traces() {
  it("should save an authentication trace in db", function() {
    const options = {
      inMemoryOnly: true
    };

    const data_store = new UserDataStore(options, nedb);
    const userid = "user";
    const type = "1stfactor";
    const is_success = false;
    return data_store.save_authentication_trace(userid, type, is_success)
    .then(function(doc) {
      assert("_id" in doc);
      assert.equal(doc.userid, "user");
      assert.equal(doc.is_success, false);
      assert.equal(doc.type, "1stfactor");
      return Promise.resolve();
    });
  });

  it("should return 3 last authentication traces", function() {
    const options = {
      inMemoryOnly: true
    };

    const data_store = new UserDataStore(options, nedb);
    const userid = "user";
    const type = "1stfactor";
    const is_success = false;
    MockDate.set("2/1/2000");
    return data_store.save_authentication_trace(userid, type, false)
    .then(function(doc) {
      MockDate.set("1/2/2000");
      return data_store.save_authentication_trace(userid, type, true);
    })
    .then(function(doc) {
      MockDate.set("1/7/2000");
      return data_store.save_authentication_trace(userid, type, false);
    })
    .then(function(doc) {
      MockDate.set("1/2/2000");
      return data_store.save_authentication_trace(userid, type, false);
    })
    .then(function(doc) {
      MockDate.set("1/5/2000");
      return data_store.save_authentication_trace(userid, type, false);
    })
    .then(function(doc) {
      return data_store.get_last_authentication_traces(userid, type, false, 3);
    })
    .then(function(docs) {
      assert.equal(docs.length, 3);
      assert.deepEqual(docs[0].date, new Date("2/1/2000"));
      assert.deepEqual(docs[1].date, new Date("1/7/2000"));
      assert.deepEqual(docs[2].date, new Date("1/5/2000"));
      return Promise.resolve();
    });
  });
}
