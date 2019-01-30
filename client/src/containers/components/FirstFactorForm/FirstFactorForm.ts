import { connect } from 'react-redux';
import { Dispatch } from 'redux';
import { authenticateFailure, authenticateSuccess, authenticate } from '../../../reducers/Portal/FirstFactor/actions';
import FirstFactorForm, { StateProps } from '../../../components/FirstFactorForm/FirstFactorForm';
import { RootState } from '../../../reducers';
import * as AutheliaService from '../../../services/AutheliaService';
import to from 'await-to-js';
import FetchStateBehavior from '../../../behaviors/FetchStateBehavior';

const mapStateToProps = (state: RootState): StateProps => {
  return {
    error: state.firstFactor.error,
    formDisabled: state.firstFactor.loading,
  };
}

function onAuthenticationRequested(dispatch: Dispatch) {
  return async (username: string, password: string, rememberMe: boolean) => {
    let err, res;
    
    // Validate first factor
    dispatch(authenticate());
    [err, res] = await to(AutheliaService.postFirstFactorAuth(username, password, rememberMe));

    if (err) {
      await dispatch(authenticateFailure(err.message));
      return;
    }

    if (!res) {
      await dispatch(authenticateFailure('No response'));
      return;
    }

    if (res.status !== 204) {
      const json = await res.json();
      if ('error' in json) {
        await dispatch(authenticateFailure(json['error']));
        return;
      }
    }
    
    dispatch(authenticateSuccess());

    // fetch state
    FetchStateBehavior(dispatch);
  }
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    onAuthenticationRequested: onAuthenticationRequested(dispatch),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(FirstFactorForm);