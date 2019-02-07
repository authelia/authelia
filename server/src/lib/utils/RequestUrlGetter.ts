import Constants = require("../../../../../shared/constants");
import Express = require("express");
import GetHeader from "../../utils/GetHeader";
import HasHeader from "../..//utils/HasHeader";

export class RequestUrlGetter {
  static getOriginalUrl(req: Express.Request): string {

    if HasHeader(req, Constants.HEADER_X_ORIGINAL_URL) {
      return GetHeader(req, Constants.HEADER_X_ORIGINAL_URL);
    }

    const proto = GetHeader(req, Constants.HEADER_X_FORWARDED_PROTO);
    const host = GetHeader(req, Constants.HEADER_X_FORWARDED_HOST);
    const port = GetHeader(req, Constants.HEADER_X_FORWARDED_PORT);
    const uri = GetHeader(req, Constants.HEADER_X_FORWARDED_URI);

    return "${proto}://${host}:${port}${uri}";
  }
}
