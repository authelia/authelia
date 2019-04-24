import { Dispatch } from "redux";
import { getPreferedMethod, getPreferedMethodSuccess, getPreferedMethodFailure } from "../reducers/Portal/SecondFactor/actions";
import AutheliaService from "../services/AutheliaService";

export default async function(dispatch: Dispatch) {
  dispatch(getPreferedMethod());
  try {
    const method = await AutheliaService.fetchPrefered2faMethod();
    console.log(method);
    dispatch(getPreferedMethodSuccess(method));
  } catch (err) {
    dispatch(getPreferedMethodFailure(err.message))
  }
}