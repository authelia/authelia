import { Dispatch } from "redux";
import { fetchStateFailure, fetchStateSuccess } from "../reducers/Portal/Authentication/actions";
import AutheliaService from "../services/AutheliaService";

export default async function(dispatch: Dispatch) {
  try {
    const state = await AutheliaService.fetchState();
    dispatch(fetchStateSuccess(state));
  } catch (err) {
    dispatch(fetchStateFailure(err.message));
  }
}