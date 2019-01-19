import { Dispatch } from "redux";
import { logout, logoutFailure, logoutSuccess } from "../reducers/Portal/SecondFactor/actions";
import to from "await-to-js";
import * as AutheliaService from '../services/AutheliaService';
import fetchState from "./FetchStateBehavior";

export default async function(dispatch: Dispatch) {
  await dispatch(logout());
  let err, res;
  [err, res] = await to(AutheliaService.postLogout());

  if (err) {
    await dispatch(logoutFailure(err.message));
    return;
  }
  await dispatch(logoutSuccess());
  await fetchState(dispatch);
}