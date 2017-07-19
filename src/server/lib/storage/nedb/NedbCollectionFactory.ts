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
  private options: NedbOptions;

  constructor(options: NedbOptions) {
    this.options = options;
  }

  build(collectionName: string): ICollection {
    const datastoreOptions = {
      inMemoryOnly: this.options.inMemoryOnly || false,
      autoload: true,
      filename: (this.options.directory) ? path.resolve(this.options.directory, collectionName) : undefined
    };

    return new NedbCollection(datastoreOptions);
  }
}