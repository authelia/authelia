import Assert = require("assert");
import Sinon = require("sinon");
import MongoDB = require("mongodb");
import BluebirdPromise = require("bluebird");
import { MongoClientStub } from "../../mocks/connectors/mongo/MongoClientStub";
import { MongoCollection } from "../../../src/lib/storage/mongo/MongoCollection";

describe("MongoCollection", function () {
  let mongoCollectionStub: any;
  let findStub: Sinon.SinonStub;
  let findOneStub: Sinon.SinonStub;
  let insertStub: Sinon.SinonStub;
  let updateStub: Sinon.SinonStub;
  let removeStub: Sinon.SinonStub;
  let countStub: Sinon.SinonStub;

  before(function () {
    mongoCollectionStub = Sinon.createStubInstance(require("mongodb").Collection as any);
    findStub = mongoCollectionStub.find as Sinon.SinonStub;
    findOneStub = mongoCollectionStub.findOne as Sinon.SinonStub;
    insertStub = mongoCollectionStub.insert as Sinon.SinonStub;
    updateStub = mongoCollectionStub.update as Sinon.SinonStub;
    removeStub = mongoCollectionStub.remove as Sinon.SinonStub;
    countStub = mongoCollectionStub.count as Sinon.SinonStub;
  });

  describe("find", function () {
    it("should find a document in the collection", function () {
      const collection = new MongoCollection(mongoCollectionStub);
      findStub.returns({
        sort: Sinon.stub().returns({
          limit: Sinon.stub().returns({
            toArray: Sinon.stub().yields(undefined, [])
          })
        })
      });

      return collection.find({ key: "KEY" })
        .then(function () {
          Assert(findStub.calledWith({ key: "KEY" }));
        });
    });
  });

  describe("findOne", function () {
    it("should find one document in the collection", function () {
      const collection = new MongoCollection(mongoCollectionStub);
      findOneStub.yields(undefined, {});

      return collection.findOne({ key: "KEY" })
        .then(function () {
          Assert(findOneStub.calledWith({ key: "KEY" }));
        });
    });
  });

  describe("insert", function () {
    it("should insert a document in the collection", function () {
      const collection = new MongoCollection(mongoCollectionStub);
      insertStub.yields(undefined, {});

      return collection.insert({ key: "KEY" })
        .then(function () {
          Assert(insertStub.calledWith({ key: "KEY" }));
        });
    });
  });

  describe("update", function () {
    it("should update a document in the collection", function () {
      const collection = new MongoCollection(mongoCollectionStub);
      updateStub.yields(undefined, {});

      return collection.update({ key: "KEY" }, { key: "KEY", value: 1 })
        .then(function () {
          Assert(updateStub.calledWith({ key: "KEY" }, { key: "KEY", value: 1 }));
        });
    });
  });

  describe("remove", function () {
    it("should remove a document in the collection", function () {
      const collection = new MongoCollection(mongoCollectionStub);
      removeStub.yields(undefined, {});

      return collection.remove({ key: "KEY" })
        .then(function () {
          Assert(removeStub.calledWith({ key: "KEY" }));
        });
    });
  });

  describe("count", function () {
    it("should count documents in the collection", function () {
      const collection = new MongoCollection(mongoCollectionStub);
      countStub.yields(undefined, {});

      return collection.count({ key: "KEY" })
        .then(function () {
          Assert(countStub.calledWith({ key: "KEY" }));
        });
    });
  });
});
