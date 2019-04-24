import { connect } from 'react-redux';
import { Dispatch } from 'redux';
import {
  authenticateFailure,
  authenticateSuccess,
  authenticate,
  setUsername,
  setPassword
} from '../../../reducers/Portal/FirstFactor/actions';
import FirstFactorForm, { StateProps, OwnProps } from '../../../components/FirstFactorForm/FirstFactorForm';
import { RootState } from '../../../reducers';
import FetchStateBehavior from '../../../behaviors/FetchStateBehavior';
import AutheliaService from '../../../services/AutheliaService';

const mapStateToProps = (state: RootState): StateProps => {
  return {
    error: state.firstFactor.error,
    formDisabled: state.firstFactor.loading,
    username: state.firstFactor.username,
    password: state.firstFactor.password,
  };
}

function onAuthenticationRequested(dispatch: Dispatch, redirectionUrl: string | null) {
  return async (username: string, password: string, rememberMe: boolean): Promise<void> => {    
    // Validate first factor
    dispatch(authenticate());
    try {
      const redirectOrUndefined = await AutheliaService.postFirstFactorAuth(
        username, password, rememberMe, redirectionUrl);
      if (redirectOrUndefined) {
        window.location.href = redirectOrUndefined.redirect;
        return;
      }
      dispatch(authenticateSuccess());
      dispatch(setUsername(''));
      dispatch(setPassword(''));
      // fetch state to move to next stage in case redirect is not possible
      await FetchStateBehavior(dispatch);
    } catch (err) {
      dispatch(setPassword(''));
      dispatch(authenticateFailure(err.message));
    }
  }
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: OwnProps) => {
  return {
    onUsernameChanged: function(username: string) {
      dispatch(setUsername(username));
    },
    onPasswordChanged: function(password: string) {
      dispatch(setPassword(password));
    },
    onAuthenticationRequested: onAuthenticationRequested(dispatch, ownProps.redirectionUrl),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(FirstFactorForm);