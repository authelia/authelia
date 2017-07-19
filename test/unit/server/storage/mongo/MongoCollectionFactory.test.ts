import Assert = require("assert");
import Sinon = require("sinon");
import { MongoClientStub } from "../../mocks/connectors/mongo/MongoClientStub";
import { MongoCollectionFactory } from "../../../../../src/server/lib/storage/mongo/MongoCollectionFactory";

describe("MongoCollectionFactory", function () {
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
