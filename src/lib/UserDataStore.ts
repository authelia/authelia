import * as Promise from "bluebird";
import * as path from "path";
import Nedb = require("nedb");
import { NedbAsync } from "nedb";
import { TOTPSecret } from "../types/TOTPSecret";

// Constants

const U2F_META_COLLECTION_NAME = "u2f_meta";
const IDENTITY_CHECK_TOKENS_COLLECTION_NAME = "identity_check_tokens";
const AUTHENTICATION_TRACES_COLLECTION_NAME = "authentication_traces";
const TOTP_SECRETS_COLLECTION_NAME = "totp_secrets";


export interface TOTPSecretDocument {
  userid: string;
  secret: TOTPSecret;
}

export interface U2FMetaDocument {
  meta: object;
  userid: string;
  appid: string;
}

export interface Options {
  inMemoryOnly?: boolean;
  directory?: string;
}


// Source

export default class UserDataStore {
  private _u2f_meta_collection: NedbAsync;
  private _identity_check_tokens_collection: NedbAsync;
  private _authentication_traces_collection: NedbAsync;
  private _totp_secret_collection: NedbAsync;

  constructor(options?: Options) {
    this._u2f_meta_collection = create_collection(U2F_META_COLLECTION_NAME, options);
    this._identity_check_tokens_collection =
      create_collection(IDENTITY_CHECK_TOKENS_COLLECTION_NAME, options);
    this._authentication_traces_collection =
      create_collection(AUTHENTICATION_TRACES_COLLECTION_NAME, options);
    this._totp_secret_collection =
      create_collection(TOTP_SECRETS_COLLECTION_NAME, options);
  }

  set_u2f_meta(userid: string, appid: string, meta: Object): Promise<any> {
    const newDocument = {
      userid: userid,
      appid: appid,
      meta: meta
    };

    const filter = {
      userid: userid,
      appid: appid
    };

    return this._u2f_meta_collection.updateAsync(filter, newDocument, { upsert: true });
  }

  get_u2f_meta(userid: string, appid: string): Promise<U2FMetaDocument> {
    const filter = {
      userid: userid,
      appid: appid
    };
    return this._u2f_meta_collection.findOneAsync(filter);
  }

  save_authentication_trace(userid: string, type: string, is_success: boolean) {
    const newDocument = {
      userid: userid,
      date: new Date(),
      is_success: is_success,
      type: type
    };

    return this._authentication_traces_collection.insertAsync(newDocument);
  }

  get_last_authentication_traces(userid: string, type: string, is_success: boolean, count: number): Promise<any> {
    const q = {
      userid: userid,
      type: type,
      is_success: is_success
    };

    const query = this._authentication_traces_collection.find(q)
      .sort({ date: -1 }).limit(count);
    const query_promisified = Promise.promisify(query.exec, { context: query });
    return query_promisified();
  }

  issue_identity_check_token(userid: string, token: string, data: string | object, max_age: number): Promise<any> {
    const newDocument = {
      userid: userid,
      token: token,
      content: {
        userid: userid,
        data: data
      },
      max_date: new Date(new Date().getTime() + max_age)
    };

    return this._identity_check_tokens_collection.insertAsync(newDocument);
  }

  consume_identity_check_token(token: string): Promise<any> {
    const query = {
      token: token
    };

    return this._identity_check_tokens_collection.findOneAsync(query)
      .then(function (doc) {
        if (!doc) {
          return Promise.reject("Registration token does not exist");
        }

        const max_date = doc.max_date;
        const current_date = new Date();
        if (current_date > max_date) {
          return Promise.reject("Registration token is not valid anymore");
        }
        return Promise.resolve(doc.content);
      })
      .then((content) => {
        return Promise.join(this._identity_check_tokens_collection.removeAsync(query),
          Promise.resolve(content));
      })
      .then((v) => {
        return Promise.resolve(v[1]);
      });
  }

  set_totp_secret(userid: string, secret: TOTPSecret): Promise<any> {
    const doc = {
      userid: userid,
      secret: secret
    };

    const query = {
      userid: userid
    };
    return this._totp_secret_collection.updateAsync(query, doc, { upsert: true });
  }

  get_totp_secret(userid: string): Promise<TOTPSecretDocument> {
    const query = {
      userid: userid
    };
    return this._totp_secret_collection.findOneAsync(query);
  }
}

function create_collection(name: string, options: any): NedbAsync {
  const datastore_options = {
    inMemoryOnly: options.inMemoryOnly || false,
    autoload: true,
    filename: ""
  };

  if (options.directory)
    datastore_options.filename = path.resolve(options.directory, name);

  return Promise.promisifyAll(new Nedb(datastore_options)) as NedbAsync;
}
