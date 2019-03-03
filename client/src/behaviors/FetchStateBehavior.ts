import { Dispatch } from "redux";
import * as AutheliaService from '../services/AutheliaService';
import { fetchStateFailure, fetchStateSuccess } from "../reducers/Portal/Authentication/actions";
import to from "await-to-js";

export default async function(dispatch: Dispatch) {
  let err, res;
  [err, res] = await to(AutheliaService.fetchState());
  if (err) {
    await dispatch(fetchStateFailure(err.message));
    return;
  }
  if (!res) {
    await dispatch(fetchStateFailure('No response'));
    return
  }
  await dispatch(fetchStateSuccess(res));
  return res;
}