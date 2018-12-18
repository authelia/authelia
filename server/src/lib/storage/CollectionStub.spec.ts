import BluebirdPromise = require("bluebird");
import Sinon = require("sinon");
import { ICollection } from "./ICollection";

export class CollectionStub implements ICollection {
    findStub: Sinon.SinonStub;
    findOneStub: Sinon.SinonStub;
    updateStub: Sinon.SinonStub;
    removeStub: Sinon.SinonStub;
    insertStub: Sinon.SinonStub;

    constructor() {
        this.findStub = Sinon.stub();
        this.findOneStub = Sinon.stub();
        this.updateStub = Sinon.stub();
        this.removeStub = Sinon.stub();
        this.insertStub = Sinon.stub();
    }

    find(filter: any, sortKeys: any, count: number): BluebirdPromise<any> {
        return this.findStub(filter, sortKeys, count);
    }

    findOne(filter: any): BluebirdPromise<any> {
        return this.findOneStub(filter);
    }

    update(filter: any, document: any, options: any): BluebirdPromise<any> {
        return this.updateStub(filter, document, options);
    }

    remove(filter: any): BluebirdPromise<any> {
        return this.removeStub(filter);
    }

    insert(document: any): BluebirdPromise<any> {
        return this.insertStub(document);
    }
}
