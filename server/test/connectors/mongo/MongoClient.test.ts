import Assert = require("assert");
import Sinon = require("sinon");
import MongoDB = require("mongodb");
import { MongoClient } from "../../../src/lib/connectors/mongo/MongoClient";

describe("MongoClient", function () {
  let mongoClientConnectStub: Sinon.SinonStub;
  let mongoDatabase: any;
  let mongoDatabaseCollectionStub: Sinon.SinonStub;

  describe("collection", function () {
    before(function () {
      mongoDatabaseCollectionStub = Sinon.stub();
      mongoDatabase = {
        collection: mongoDatabaseCollectionStub
      };

      mongoClientConnectStub = Sinon.stub(MongoDB.MongoClient, "connect");
      mongoClientConnectStub.yields(undefined, mongoDatabase);
    });

    after(function () {
      mongoClientConnectStub.restore();
    });

    it("should create a collection", function () {
      const COLLECTION_NAME = "mycollection";
      const client = new MongoClient(mongoDatabase);

      mongoDatabaseCollectionStub.returns({});

      const collection = client.collection(COLLECTION_NAME);

      Assert(collection);
      Assert(mongoDatabaseCollectionStub.calledWith(COLLECTION_NAME ));
    });
  });
});
