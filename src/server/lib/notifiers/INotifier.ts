
import * as BluebirdPromise from "bluebird";
import { Identity } from "../../../types/Identity";

export abstract class INotifier {
  abstract notify(identity: Identity, subject: string, link: string): BluebirdPromise<void>;
}