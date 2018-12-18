import Sinon = require("sinon");
import Assert = require("assert");

import { NedbCollectionFactory } from "./NedbCollectionFactory";

describe("storage/nedb/NedbCollectionFactory", function() {
  it("should create a nedb collection", function() {
    const nedbOptions = {
      inMemoryOnly: true
    };
    const factory = new NedbCollectionFactory(nedbOptions);

    const collection = factory.build("mycollection");
    Assert(collection);
  });
});