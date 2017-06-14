
import * as assert from "assert";
import * as Promise from "bluebird";
import * as sinon from "sinon";
import * as MockDate from "mockdate";
import UserDataStore from "../../../src/server/lib/UserDataStore";
import nedb = require("nedb");

describe("test user data store", function() {
  describe("test totp secrets store", test_totp_secrets);
});

function test_totp_secrets() {
  it("should save and reload a totp secret", function() {
    const options = {
      inMemoryOnly: true
    };

    const data_store = new UserDataStore(options, nedb);
    const userid = "user";
    const secret = {
      ascii: "abc",
      base32: "ABCDKZLEFZGREJK",
      otpauth_url: "totp://test"
    };

    return data_store.set_totp_secret(userid, secret)
    .then(function() {
      return data_store.get_totp_secret(userid);
    })
    .then(function(doc) {
      assert("_id" in doc);
      assert.equal(doc.userid, "user");
      assert.equal(doc.secret.ascii, "abc");
      assert.equal(doc.secret.base32, "ABCDKZLEFZGREJK");
      return Promise.resolve();
    });
  });

  it("should only remember last secret", function() {
    const options = {
      inMemoryOnly: true
    };

    const data_store = new UserDataStore(options, nedb);
    const userid = "user";
    const secret1 = {
      ascii: "abc",
      base32: "ABCDKZLEFZGREJK",
      otpauth_url: "totp://test"
    };
    const secret2 = {
      ascii: "def",
      base32: "XYZABC",
      otpauth_url: "totp://test"
    };

    return data_store.set_totp_secret(userid, secret1)
    .then(function() {
      return data_store.set_totp_secret(userid, secret2);
    })
    .then(function() {
      return data_store.get_totp_secret(userid);
    })
    .then(function(doc) {
      assert("_id" in doc);
      assert.equal(doc.userid, "user");
      assert.equal(doc.secret.ascii, "def");
      assert.equal(doc.secret.base32, "XYZABC");
      return Promise.resolve();
    });
  });
}
