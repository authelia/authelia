
import { ICollection } from "./ICollection";

export interface ICollectionFactory {
    build(collectionName: string): ICollection;
}