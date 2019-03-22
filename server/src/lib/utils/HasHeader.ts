import * as Express from "express";
import * as ObjectPath from "object-path";

/**
 *
 * @param req The express request to extract headers from
 * @param header The name of the header to check the existence of.
 * @returns true if the header is found, otherwise false.
 */
export default function(req: Express.Request, header: string): boolean {
  return ObjectPath.has<Express.Request>(req, "headers." + header);
}