
import * as BluebirdPromise from "bluebird";
import { Identity } from "../../../types/Identity";

export interface INotifier {
  notify(identity: Identity, subject: string, link: string): BluebirdPromise<void>;
}