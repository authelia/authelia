import * as Express from "express";
import * as ObjectPath from "object-path";
import { ServerVariables } from "../ServerVariables";
import { GET_VARIABLE_KEY } from "../constants";

/**
 *
 * @param req The express request to extract headers from
 * @param header The name of the header to extract in lowercase.
 * @returns The header if found, otherwise undefined.
 */
export default function(req: Express.Request, header: string): string | undefined {
  const variables: ServerVariables = req.app.get(GET_VARIABLE_KEY);
  if (!variables) throw new Error("There are no server variables set.");

  const value = ObjectPath.get<Express.Request, string>(req, "headers." + header, undefined);
  variables.logger.debug(req, "Header %s is set to %s", header, value);
  return value;
}