
import * as BluebirdPromise from "bluebird";
import exceptions = require("./Exceptions");

const REGULATION_TRACE_TYPE = "regulation";
const MAX_AUTHENTICATION_COUNT_IN_TIME_RANGE = 3;

interface DatedDocument {
  date: Date;
}

export class AuthenticationRegulator {
  private _user_data_store: any;
  private _lock_time_in_seconds: number;

  constructor(user_data_store: any, lock_time_in_seconds: number) {
    this._user_data_store = user_data_store;
    this._lock_time_in_seconds = lock_time_in_seconds;
  }

  // Mark authentication
  mark(userid: string, is_success: boolean): BluebirdPromise<void> {
    return this._user_data_store.save_authentication_trace(userid, REGULATION_TRACE_TYPE, is_success);
  }

  regulate(userid: string): BluebirdPromise<void> {
    return this._user_data_store.get_last_authentication_traces(userid, REGULATION_TRACE_TYPE, false, 3)
    .then((docs: Array<DatedDocument>) => {
      if (docs.length < MAX_AUTHENTICATION_COUNT_IN_TIME_RANGE) {
        // less than the max authorized number of authentication in time range, thus authorizing access
        return BluebirdPromise.resolve();
      }

      const oldest_doc = docs[MAX_AUTHENTICATION_COUNT_IN_TIME_RANGE - 1];
      const no_lock_min_date = new Date(new Date().getTime() - this._lock_time_in_seconds * 1000);
      if (oldest_doc.date > no_lock_min_date) {
        throw new exceptions.AuthenticationRegulationError("Max number of authentication. Please retry in few minutes.");
      }

      return BluebirdPromise.resolve();
    });
  }
}
