import Sinon = require("sinon");
import Assert = require("assert");

import { NedbCollection } from "./NedbCollection";

describe("storage/nedb/NedbCollection", function () {
  describe("insert", function () {
    it("should insert one entry", function () {
      const nedbOptions = {
        inMemoryOnly: true
      };
      const collection = new NedbCollection(nedbOptions);

      collection.insert({ key: "coucou" });

      return collection.count({}).then(function (count: number) {
        Assert.equal(1, count);
      });
    });

    it("should insert three entries", function () {
      const nedbOptions = {
        inMemoryOnly: true
      };
      const collection = new NedbCollection(nedbOptions);

      collection.insert({ key: "coucou" });
      collection.insert({ key: "hello" });
      collection.insert({ key: "hey" });

      return collection.count({}).then(function (count: number) {
        Assert.equal(3, count);
      });
    });
  });

  describe("find", function () {
    let collection: NedbCollection;
    before(function () {
      const nedbOptions = {
        inMemoryOnly: true
      };
      collection = new NedbCollection(nedbOptions);

      collection.insert({ key: "coucou", value: 1 });
      collection.insert({ key: "hello" });
      collection.insert({ key: "hey" });
      collection.insert({ key: "coucou", value: 2 });
    });

    it("should find one hello", function () {
      return collection.find({ key: "hello" }, { key: 1 })
        .then(function (docs: { key: string }[]) {
          Assert.equal(1, docs.length);
          Assert(docs[0].key == "hello");
        });
    });

    it("should find two coucou", function () {
      return collection.find({ key: "coucou" }, { value: 1 })
        .then(function (docs: { value: number }[]) {
          Assert.equal(2, docs.length);
        });
    });
  });

  describe("findOne", function () {
    let collection: NedbCollection;
    before(function () {
      const nedbOptions = {
        inMemoryOnly: true
      };
      collection = new NedbCollection(nedbOptions);

      collection.insert({ key: "coucou", value: 1 });
      collection.insert({ key: "coucou", value: 1 });
      collection.insert({ key: "coucou", value: 1 });
      collection.insert({ key: "coucou", value: 1 });
    });

    it("should find two coucou", function () {
      const doc = { key: "coucou", value: 1 };
      return collection.count(doc)
        .then(function (count: number) {
          Assert.equal(4, count);
          return collection.findOne(doc);
        });
    });
  });

  describe("update", function () {
    let collection: NedbCollection;
    before(function () {
      const nedbOptions = {
        inMemoryOnly: true
      };
      collection = new NedbCollection(nedbOptions);

      collection.insert({ key: "coucou", value: 1 });
    });

    it("should update the value", function () {
      return collection.update({ key: "coucou" }, { key: "coucou", value: 2 }, { multi: true })
        .then(function () {
          return collection.find({ key: "coucou" });
        })
        .then(function (docs: { key: string, value: number }[]) {
          Assert.equal(1, docs.length);
          Assert.equal(2, docs[0].value);
        });
    });
  });

  describe("update", function () {
    let collection: NedbCollection;
    before(function () {
      const nedbOptions = {
        inMemoryOnly: true
      };
      collection = new NedbCollection(nedbOptions);

      collection.insert({ key: "coucou" });
      collection.insert({ key: "hello" });
    });

    it("should update the value", function () {
      return collection.remove({ key: "coucou" })
        .then(function () {
          return collection.count({});
        })
        .then(function (count: number) {
          Assert.equal(1, count);
        });
    });
  });
});