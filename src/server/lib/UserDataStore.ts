import * as BluebirdPromise from "bluebird";
import * as path from "path";
import { NedbAsync } from "nedb";
import { TOTPSecret } from "../../types/TOTPSecret";
import { Nedb } from "../../types/Dependencies";
import u2f = require("u2f");

// Constants

const U2F_META_COLLECTION_NAME = "u2f_meta";
const IDENTITY_CHECK_TOKENS_COLLECTION_NAME = "identity_check_tokens";
const AUTHENTICATION_TRACES_COLLECTION_NAME = "authentication_traces";
const TOTP_SECRETS_COLLECTION_NAME = "totp_secrets";


export interface TOTPSecretDocument {
  userid: string;
  secret: TOTPSecret;
}

export interface U2FRegistrationDocument {
  keyHandle: string;
  publicKey: string;
  userId: string;
  appId: string;
}

export interface Options {
  inMemoryOnly?: boolean;
  directory?: string;
}

export interface IdentityValidationRequestContent {
  userid: string;
  data: string;
}

export interface IdentityValidationRequestDocument {
  userid: string;
  token: string;
  content: IdentityValidationRequestContent;
  max_date: Date;
}

interface U2FRegistrationFilter {
  userId: string;
  appId: string;
}

// Source

export default class UserDataStore {
  private _u2f_meta_collection: NedbAsync;
  private _identity_check_tokens_collection: NedbAsync;
  private _authentication_traces_collection: NedbAsync;
  private _totp_secret_collection: NedbAsync;
  private nedb: Nedb;

  constructor(options: Options, nedb: Nedb) {
    this.nedb = nedb;
    this._u2f_meta_collection = this.create_collection(U2F_META_COLLECTION_NAME, options);
    this._identity_check_tokens_collection =
      this.create_collection(IDENTITY_CHECK_TOKENS_COLLECTION_NAME, options);
    this._authentication_traces_collection =
      this.create_collection(AUTHENTICATION_TRACES_COLLECTION_NAME, options);
    this._totp_secret_collection =
      this.create_collection(TOTP_SECRETS_COLLECTION_NAME, options);
  }

  set_u2f_meta(userId: string, appId: string, keyHandle: string, publicKey: string): BluebirdPromise<any> {
    const newDocument: U2FRegistrationDocument = {
      userId: userId,
      appId: appId,
      keyHandle: keyHandle,
      publicKey: publicKey
    };

    const filter: U2FRegistrationFilter = {
      userId: userId,
      appId: appId
    };

    return this._u2f_meta_collection.updateAsync(filter, newDocument, { upsert: true });
  }

  get_u2f_meta(userId: string, appId: string): BluebirdPromise<U2FRegistrationDocument> {
    const filter: U2FRegistrationFilter = {
      userId: userId,
      appId: appId
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

  get_last_authentication_traces(userid: string, type: string, is_success: boolean, count: number): BluebirdPromise<any> {
    const q = {
      userid: userid,
      type: type,
      is_success: is_success
    };

    const query = this._authentication_traces_collection.find(q)
      .sort({ date: -1 }).limit(count);
    const query_promisified = BluebirdPromise.promisify(query.exec, { context: query });
    return query_promisified();
  }

  issue_identity_check_token(userid: string, token: string, data: string | object, max_age: number): BluebirdPromise<any> {
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

  consume_identity_check_token(token: string): BluebirdPromise<IdentityValidationRequestContent> {
    const query = {
      token: token
    };

    return this._identity_check_tokens_collection.findOneAsync(query)
      .then(function (doc) {
        if (!doc) {
          return BluebirdPromise.reject(new Error("Registration token does not exist"));
        }

        const max_date = doc.max_date;
        const current_date = new Date();
        if (current_date > max_date)
          return BluebirdPromise.reject(new Error("Registration token is not valid anymore"));

        return BluebirdPromise.resolve(doc.content);
      })
      .then((content) => {
        return BluebirdPromise.join(this._identity_check_tokens_collection.removeAsync(query),
          BluebirdPromise.resolve(content));
      })
      .then((v) => {
        return BluebirdPromise.resolve(v[1]);
      });
  }

  set_totp_secret(userid: string, secret: TOTPSecret): BluebirdPromise<any> {
    const doc = {
      userid: userid,
      secret: secret
    };

    const query = {
      userid: userid
    };
    return this._totp_secret_collection.updateAsync(query, doc, { upsert: true });
  }

  get_totp_secret(userid: string): BluebirdPromise<TOTPSecretDocument> {
    const query = {
      userid: userid
    };
    return this._totp_secret_collection.findOneAsync(query);
  }

  private create_collection(name: string, options: any): NedbAsync {
    const datastore_options = {
      inMemoryOnly: options.inMemoryOnly || false,
      autoload: true,
      filename: ""
    };

    if (options.directory)
      datastore_options.filename = path.resolve(options.directory, name);

    return BluebirdPromise.promisifyAll(new this.nedb(datastore_options)) as NedbAsync;
  }
}
