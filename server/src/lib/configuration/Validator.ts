import Ajv = require("ajv");
import Path = require("path");

export class Validator {
  static isValid(configuration: any) {
    const schema = require(Path.resolve(__dirname, "./Configuration.schema.json"));
    const ajv = new Ajv({
      allErrors: true,
      missingRefs: "fail"
    });
    ajv.addMetaSchema(require("ajv/lib/refs/json-schema-draft-04.json"));
    const valid = ajv.validate(schema, configuration);
    if (!valid) {
      for (const i in ajv.errors) {
        console.log(ajv.errorsText([ajv.errors[i]]));
      }
    }
    return valid;
  }
}