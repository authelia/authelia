import U2f = require("u2f");

export interface SignMessage {
  request: U2f.Request;
  keyHandle: string;
}