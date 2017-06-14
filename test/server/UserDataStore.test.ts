
import UserDataStore from "../../src/server/lib/UserDataStore";
import { U2FRegistrationDocument, Options } from "../../src/server/lib/UserDataStore";

import nedb = require("nedb");
import assert = require("assert");
import Promise = require("bluebird");
import sinon = require("sinon");
import MockDate = require("mockdate");

describe("test user data store", () => {
  let options: Options;

  beforeEach(function () {
    options = {
      inMemoryOnly: true
    };
  });


  describe("test u2f meta", () => {
    it("should save a u2f meta", function () {
      const data_store = new UserDataStore(options, nedb);

      const userid = "user";
      const app_id = "https://localhost";
      const keyHandle = "keyhandle";
      const publicKey = "publicKey";

      return data_store.set_u2f_meta(userid, app_id, keyHandle, publicKey)
        .then(function (numUpdated) {
          assert.equal(1, numUpdated);
          return Promise.resolve();
        });
    });

    it("should retrieve no u2f meta", function () {
      const options = {
        inMemoryOnly: true
      };

      const data_store = new UserDataStore(options, nedb);

      const userid = "user";
      const app_id = "https://localhost";
      const meta = {
        publicKey: "pbk"
      };

      return data_store.get_u2f_meta(userid, app_id)
        .then(function (doc) {
          assert.equal(undefined, doc);
          return Promise.resolve();
        });
    });

    it("should insert and retrieve a u2f meta", function () {
      const options = {
        inMemoryOnly: true
      };

      const data_store = new UserDataStore(options, nedb);

      const userid = "user";
      const app_id = "https://localhost";
      const keyHandle = "keyHandle";
      const publicKey = "publicKey";

      return data_store.set_u2f_meta(userid, app_id, keyHandle, publicKey)
        .then(function (numUpdated: number) {
          assert.equal(1, numUpdated);
          return data_store.get_u2f_meta(userid, app_id);
        })
        .then(function (doc: U2FRegistrationDocument) {
          assert.deepEqual(keyHandle, doc.keyHandle);
          assert.deepEqual(publicKey, doc.publicKey);
          assert.deepEqual(userid, doc.userId);
          assert.deepEqual(app_id, doc.appId);
          assert("_id" in doc);
          return Promise.resolve();
        });
    });
  });


  describe("test u2f registration token", () => {
    it("should save u2f registration token", function () {
      const data_store = new UserDataStore(options, nedb);

      const userid = "user";
      const token = "token";
      const max_age = 60;
      const content = "abc";

      return data_store.issue_identity_check_token(userid, token, content, max_age)
        .then(function (document) {
          assert.equal(document.userid, userid);
          assert.equal(document.token, token);
          assert.deepEqual(document.content, { userid: "user", data: content });
          assert("max_date" in document);
          assert("_id" in document);
          return Promise.resolve();
        })
        .catch(function (err) {
          console.error(err);
          return Promise.reject(err);
        });
    });

    it("should save u2f registration token and consume it", function (done) {
      const data_store = new UserDataStore(options, nedb);

      const userid = "user";
      const token = "token";
      const max_age = 50;

      data_store.issue_identity_check_token(userid, token, {}, max_age)
        .then(function (document) {
          return data_store.consume_identity_check_token(token);
        })
        .then(function () {
          done();
        })
        .catch(function (err) {
          console.error(err);
        });
    });

    it("should not be able to consume registration token twice", function (done) {
      const data_store = new UserDataStore(options, nedb);

      const userid = "user";
      const token = "token";
      const max_age = 50;

      data_store.issue_identity_check_token(userid, token, {}, max_age)
        .then(function (document) {
          return data_store.consume_identity_check_token(token);
        })
        .then(function (document) {
          return data_store.consume_identity_check_token(token);
        })
        .catch(function (err) {
          console.error(err);
          done();
        });
    });

    it("should fail when token does not exist", function () {
      const data_store = new UserDataStore(options, nedb);

      const token = "token";

      return data_store.consume_identity_check_token(token)
        .then(function (document) {
          return Promise.reject("Error while checking token");
        })
        .catch(function (err) {
          return Promise.resolve(err);
        });
    });

    it("should fail when token expired", function (done) {
      const data_store = new UserDataStore(options, nedb);

      const userid = "user";
      const token = "token";
      const max_age = 60;
      MockDate.set("1/1/2000");

      data_store.issue_identity_check_token(userid, token, {}, max_age)
        .then(function () {
          MockDate.set("1/2/2000");
          return data_store.consume_identity_check_token(token);
        })
        .catch(function (err) {
          MockDate.reset();
          done();
        });
    });

    it("should save the userid and some data with the token", function (done) {
      const data_store = new UserDataStore(options, nedb);

      const userid = "user";
      const token = "token";
      const max_age = 60;
      MockDate.set("1/1/2000");
      const data = "abc";

      data_store.issue_identity_check_token(userid, token, data, max_age)
        .then(function () {
          return data_store.consume_identity_check_token(token);
        })
        .then(function (content) {
          const expected_content = {
            userid: "user",
            data: "abc"
          };
          assert.deepEqual(content, expected_content);
          done();
        });
    });
  });
});
