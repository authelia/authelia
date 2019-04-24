import { Dispatch } from "redux";
import { logout, logoutFailure, logoutSuccess } from "../reducers/Portal/SecondFactor/actions";
import fetchState from "./FetchStateBehavior";
import AutheliaService from "../services/AutheliaService";

export default async function(dispatch: Dispatch) {
  try {
    dispatch(logout());
    await AutheliaService.postLogout();
    dispatch(logoutSuccess());
    await fetchState(dispatch);
  } catch (err) {
    dispatch(logoutFailure(err.message));
  }
}