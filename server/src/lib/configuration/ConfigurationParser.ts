
import * as ObjectPath from "object-path";
import { Configuration, complete } from "./schema/Configuration";
import Ajv = require("ajv");
import Path = require("path");
import Util = require("util");

export class ConfigurationParser {
  private static parseTypes(configuration: Configuration): string[] {
    const schema = require(Path.resolve(__dirname, "./Configuration.schema.json"));
    const ajv = new Ajv({
      allErrors: true,
      missingRefs: "fail"
    });
    ajv.addMetaSchema(require("ajv/lib/refs/json-schema-draft-06.json"));
    const valid = ajv.validate(schema, configuration);
    if (!valid)
      return ajv.errors.map(
        (e: Ajv.ErrorObject) => { return ajv.errorsText([e]); });
    return [];
  }

  static parse(configuration: Configuration): Configuration {
    const validationErrors = this.parseTypes(configuration);
    if (validationErrors.length > 0) {
      validationErrors.forEach((e: string) => { console.log(e); });
      throw new Error("Malformed configuration (schema). Please double-check your configuration file.");
    }

    const [newConfiguration, completionErrors] = complete(configuration);

    if (completionErrors.length > 0) {
      completionErrors.forEach((e: string) => { console.log(e); });
      throw new Error("Malformed configuration (validator). Please double-check your configuration file.");
    }
    return newConfiguration;
  }
}

