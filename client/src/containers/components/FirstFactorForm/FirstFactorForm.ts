import { connect } from 'react-redux';
import { Dispatch } from 'redux';
import { authenticateFailure, authenticateSuccess, authenticate } from '../../../reducers/Portal/FirstFactor/actions';
import FirstFactorForm, { StateProps, OwnProps } from '../../../components/FirstFactorForm/FirstFactorForm';
import { RootState } from '../../../reducers';
import * as AutheliaService from '../../../services/AutheliaService';
import to from 'await-to-js';
import FetchStateBehavior from '../../../behaviors/FetchStateBehavior';
import SafelyRedirectBehavior from '../../../behaviors/SafelyRedirectBehavior';

const mapStateToProps = (state: RootState): StateProps => {
  return {
    error: state.firstFactor.error,
    formDisabled: state.firstFactor.loading,
  };
}

function onAuthenticationRequested(dispatch: Dispatch, redirectionUrl: string | null) {
  return async (username: string, password: string, rememberMe: boolean) => {
    let err, res;
    
    // Validate first factor
    dispatch(authenticate());
    [err, res] = await to(AutheliaService.postFirstFactorAuth(
      username, password, rememberMe, redirectionUrl));

    if (err) {
      await dispatch(authenticateFailure(err.message));
      return;
    }

    if (!res) {
      await dispatch(authenticateFailure('No response'));
      return;
    }

    if (res.status === 200) {
      const json = await res.json();
      if ('error' in json) {
        await dispatch(authenticateFailure(json['error']));
        return;
      }

      if ('redirect' in json) {
        window.location.href = json['redirect'];
        return;
      }
    } else if (res.status === 204) {
      dispatch(authenticateSuccess());

      // fetch state to move to next stage
      FetchStateBehavior(dispatch);
    } else {
      dispatch(authenticateFailure('Unknown error'));
    }
  }
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: OwnProps) => {
  return {
    onAuthenticationRequested: onAuthenticationRequested(dispatch, ownProps.redirectionUrl),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(FirstFactorForm);