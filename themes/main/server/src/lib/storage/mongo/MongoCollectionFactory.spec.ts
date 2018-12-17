import Assert = require("assert");
import Sinon = require("sinon");
import { MongoClientStub } from "../../connectors/mongo/MongoClientStub.spec";
import { MongoCollectionFactory } from "./MongoCollectionFactory";

describe("storage/mongo/MongoCollectionFactory", function () {
  let mongoClient: MongoClientStub;

  before(function() {
    mongoClient = new MongoClientStub();
  });

  describe("create", function () {
    it("should create a collection", function () {
      const COLLECTION_NAME = "COLLECTION_NAME";

      const factory = new MongoCollectionFactory(mongoClient);
      Assert(factory.build(COLLECTION_NAME));
    });
  });
});
