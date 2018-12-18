import Assert = require("assert");
import Sinon = require("sinon");
import MongoDB = require("mongodb");
import BluebirdPromise = require("bluebird");
import { MongoClientStub } from "../../connectors/mongo/MongoClientStub.spec";
import { MongoCollection } from "./MongoCollection";

describe("storage/mongo/MongoCollection", function () {
  let mongoCollectionStub: any;
  let mongoClientStub: MongoClientStub;
  let findStub: Sinon.SinonStub;
  let findOneStub: Sinon.SinonStub;
  let insertOneStub: Sinon.SinonStub;
  let updateStub: Sinon.SinonStub;
  let removeStub: Sinon.SinonStub;
  let countStub: Sinon.SinonStub;
  const COLLECTION_NAME = "collection";

  before(function () {
    mongoClientStub = new MongoClientStub();
    mongoCollectionStub = Sinon.createStubInstance(require("mongodb").Collection as any);
    findStub = mongoCollectionStub.find as Sinon.SinonStub;
    findOneStub = mongoCollectionStub.findOne as Sinon.SinonStub;
    insertOneStub = mongoCollectionStub.insertOne as Sinon.SinonStub;
    updateStub = mongoCollectionStub.update as Sinon.SinonStub;
    removeStub = mongoCollectionStub.remove as Sinon.SinonStub;
    countStub = mongoCollectionStub.count as Sinon.SinonStub;
    mongoClientStub.collectionStub.returns(
      BluebirdPromise.resolve(mongoCollectionStub)
    );
  });

  describe("find", function () {
    it("should find a document in the collection", function () {
      const collection = new MongoCollection(COLLECTION_NAME, mongoClientStub);
      findStub.returns({
        sort: Sinon.stub().returns({
          limit: Sinon.stub().returns({
            toArray: Sinon.stub().returns(BluebirdPromise.resolve([]))
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
      const collection = new MongoCollection(COLLECTION_NAME, mongoClientStub);
      findOneStub.returns(BluebirdPromise.resolve({}));

      return collection.findOne({ key: "KEY" })
        .then(function () {
          Assert(findOneStub.calledWith({ key: "KEY" }));
        });
    });
  });

  describe("insert", function () {
    it("should insert a document in the collection", function () {
      const collection = new MongoCollection(COLLECTION_NAME, mongoClientStub);
      insertOneStub.returns(BluebirdPromise.resolve({}));

      return collection.insert({ key: "KEY" })
        .then(function () {
          Assert(insertOneStub.calledWith({ key: "KEY" }));
        });
    });
  });

  describe("update", function () {
    it("should update a document in the collection", function () {
      const collection = new MongoCollection(COLLECTION_NAME, mongoClientStub);
      updateStub.returns(BluebirdPromise.resolve({}));

      return collection.update({ key: "KEY" }, { key: "KEY", value: 1 })
        .then(function () {
          Assert(updateStub.calledWith({ key: "KEY" }, { key: "KEY", value: 1 }));
        });
    });
  });

  describe("remove", function () {
    it("should remove a document in the collection", function () {
      const collection = new MongoCollection(COLLECTION_NAME, mongoClientStub);
      removeStub.returns(BluebirdPromise.resolve({}));

      return collection.remove({ key: "KEY" })
        .then(function () {
          Assert(removeStub.calledWith({ key: "KEY" }));
        });
    });
  });

  describe("count", function () {
    it("should count documents in the collection", function () {
      const collection = new MongoCollection(COLLECTION_NAME, mongoClientStub);
      countStub.returns(BluebirdPromise.resolve({}));

      return collection.count({ key: "KEY" })
        .then(function () {
          Assert(countStub.calledWith({ key: "KEY" }));
        });
    });
  });
});
