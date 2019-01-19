import { connect } from 'react-redux';
import AuthenticationView, {StateProps, Stage, DispatchProps} from '../../../views/AuthenticationView/AuthenticationView';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import AuthenticationLevel from '../../../types/AuthenticationLevel';
import FetchStateBehavior from '../../../behaviors/FetchStateBehavior';
import { setRedirectionUrl } from '../../../reducers/Portal/Authentication/actions';

function authenticationLevelToStage(level: AuthenticationLevel): Stage {
  switch (level)  {
    case AuthenticationLevel.NOT_AUTHENTICATED:
      return Stage.FIRST_FACTOR;
    case AuthenticationLevel.ONE_FACTOR:
      return Stage.SECOND_FACTOR;
    case AuthenticationLevel.TWO_FACTOR:
      return Stage.ALREADY_AUTHENTICATED;
  }
}

const mapStateToProps = (state: RootState): StateProps => {
  const stage = (state.authentication.remoteState)
    ? authenticationLevelToStage(state.authentication.remoteState.authentication_level)
    : Stage.FIRST_FACTOR;
  return {
    redirectionUrl: state.authentication.redirectionUrl,
    remoteState: state.authentication.remoteState,
    stage: stage,
  };
}

const mapDispatchToProps = (dispatch: Dispatch): DispatchProps => {
  return {
    onInit: async (redirectionUrl?: string) => {
      await FetchStateBehavior(dispatch);
      if (redirectionUrl) {
        await dispatch(setRedirectionUrl(redirectionUrl));
      }
    }
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(AuthenticationView);