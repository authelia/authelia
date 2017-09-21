import { ICollection } from "../ICollection";
import { ICollectionFactory } from "../ICollectionFactory";
import { NedbCollection } from "./NedbCollection";
import path = require("path");
import Nedb = require("nedb");

export interface NedbOptions {
  inMemoryOnly?: boolean;
  directory?: string;
}

export class NedbCollectionFactory implements ICollectionFactory {
  private options: Nedb.DataStoreOptions;

  constructor(options: Nedb.DataStoreOptions) {
    this.options = options;
  }

  build(collectionName: string): ICollection {
    const datastoreOptions: Nedb.DataStoreOptions = {
      inMemoryOnly: this.options.inMemoryOnly || false,
      autoload: true,
      filename: (this.options.filename) ? path.resolve(this.options.filename, collectionName) : undefined
    };

    return new NedbCollection(datastoreOptions);
  }
}