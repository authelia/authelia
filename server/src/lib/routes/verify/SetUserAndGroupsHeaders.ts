import * as Express from "express";
import { HEADER_REMOTE_USER, HEADER_REMOTE_GROUPS } from "../../constants";

export default function(res: Express.Response, username: string | undefined, groups: string[] | undefined) {
  if (username) res.setHeader(HEADER_REMOTE_USER, username);
  if (groups instanceof Array) res.setHeader(HEADER_REMOTE_GROUPS, groups.join(","));
}