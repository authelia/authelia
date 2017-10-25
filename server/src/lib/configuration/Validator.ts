import Ajv = require("ajv");
import Path = require("path");
import Util = require("util");
import {
  UserConfiguration, StorageConfiguration,
  NotifierConfiguration
} from "./Configuration";

function validateSchema(configuration: UserConfiguration): string[] {
  const schema = require(Path.resolve(__dirname, "./Configuration.schema.json"));
  const ajv = new Ajv({
    allErrors: true,
    missingRefs: "fail"
  });
  ajv.addMetaSchema(require("ajv/lib/refs/json-schema-draft-04.json"));
  const valid = ajv.validate(schema, configuration);
  if (!valid)
    return ajv.errors.map(
      (e: Ajv.ErrorObject) => { return ajv.errorsText([e]); });
  return [];
}

function validateUnknownKeys(path: string, obj: any, knownKeys: string[]) {
  const keysSet = new Set(Object.keys(obj));
  const knownKeysSet = new Set(knownKeys);

  const unknownKeysSet = new Set(
    [...keysSet].filter(x => !knownKeysSet.has(x)));

  if (unknownKeysSet.size > 0) {
    const unknownKeys = Array.from(unknownKeysSet);
    return unknownKeys.map((k: string) => { return Util.format("data.%s has unknown key '%s'", path, k); });
  }
  return [];
}

function validateStorage(storage: any) {
  const ERROR = "Storage must be either 'local' or 'mongo'";

  if (!storage)
    return [];

  const errors = validateUnknownKeys("storage", storage, ["local", "mongo"]);
  if (errors.length > 0)
    return errors;

  if (storage.local && storage.mongo)
    return [ERROR];

  if (!storage.local && !storage.mongo)
    return [ERROR];

  return [];
}

function validateNotifier(notifier: NotifierConfiguration) {
  const ERROR = "Notifier must be either 'filesystem', 'email' or 'smtp'";

  if (!notifier)
    return [];

  const errors = validateUnknownKeys("notifier", notifier, ["filesystem", "email", "smtp"]);
  if (errors.length > 0)
    return errors;

  if (notifier && notifier.filesystem && notifier.email && notifier.smtp)
    return [ERROR];

  if (notifier && !notifier.filesystem && !notifier.email && !notifier.smtp)
    return [ERROR];

  return [];
}

export class Validator {
  static isValid(configuration: any): string[] {
    const schemaErrors = validateSchema(configuration);
    const storageErrors = validateStorage(configuration.storage);
    const notifierErrors = validateNotifier(configuration.notifier);

    return schemaErrors
      .concat(storageErrors)
      .concat(notifierErrors);
  }
}