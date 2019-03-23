import Sinon = require("sinon");
import { ICollection } from "./ICollection";
import { ICollectionFactory } from "./ICollectionFactory";

export class CollectionFactoryStub implements ICollectionFactory {
    buildStub: Sinon.SinonStub;

    constructor() {
        this.buildStub = Sinon.stub();
    }

    build(collectionName: string): ICollection {
        return this.buildStub(collectionName);
    }
}
