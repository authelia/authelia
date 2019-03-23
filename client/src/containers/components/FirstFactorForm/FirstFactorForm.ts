import { connect } from 'react-redux';
import { Dispatch } from 'redux';
import { authenticateFailure, authenticateSuccess, authenticate } from '../../../reducers/Portal/FirstFactor/actions';
import FirstFactorForm, { StateProps, OwnProps } from '../../../components/FirstFactorForm/FirstFactorForm';
import { RootState } from '../../../reducers';
import to from 'await-to-js';
import FetchStateBehavior from '../../../behaviors/FetchStateBehavior';
import AutheliaService from '../../../services/AutheliaService';

const mapStateToProps = (state: RootState): StateProps => {
  return {
    error: state.firstFactor.error,
    formDisabled: state.firstFactor.loading,
  };
}

function onAuthenticationRequested(dispatch: Dispatch, redirectionUrl: string | null) {
  return async (username: string, password: string, rememberMe: boolean): Promise<void> => {
    let err, res;
    
    // Validate first factor
    dispatch(authenticate());
    [err, res] = await to(AutheliaService.postFirstFactorAuth(
      username, password, rememberMe, redirectionUrl));

    if (err) {
      await dispatch(authenticateFailure(err.message));
      throw new Error(err.message);
    }

    if (!res) {
      await dispatch(authenticateFailure('No response'));
      throw new Error('No response');
    }

    if (res.status === 200) {
      const json = await res.json();
      if ('error' in json) {
        await dispatch(authenticateFailure(json['error']));
        throw new Error(json['error']);
      }

      dispatch(authenticateSuccess());
      if ('redirect' in json) {
        window.location.href = json['redirect'];
        return;
      }

      // fetch state to move to next stage in case redirect is not possible
      await FetchStateBehavior(dispatch);
    } else if (res.status === 204) {
      dispatch(authenticateSuccess());

      // fetch state to move to next stage
      await FetchStateBehavior(dispatch);
    } else {
      dispatch(authenticateFailure('Unknown error'));
      throw new Error('Unknown error... (' + res.status + ')');
    }
  }
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: OwnProps) => {
  return {
    onAuthenticationRequested: onAuthenticationRequested(dispatch, ownProps.redirectionUrl),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(FirstFactorForm);