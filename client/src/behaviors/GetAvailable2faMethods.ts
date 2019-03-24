import { Dispatch } from "redux";
import AutheliaService from "../services/AutheliaService";
import { getAvailbleMethods, getAvailbleMethodsSuccess, getAvailbleMethodsFailure } from "../reducers/Portal/SecondFactor/actions";

export default async function(dispatch: Dispatch) {
  dispatch(getAvailbleMethods());
  try {
    const methods = await AutheliaService.getAvailable2faMethods();
    dispatch(getAvailbleMethodsSuccess(methods));
  } catch (err) {
    dispatch(getAvailbleMethodsFailure(err.message))
  }
}