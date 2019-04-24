import { connect } from 'react-redux';
import QueryString from 'query-string';
import AuthenticationView, {StateProps, Stage, OwnProps} from '../../../views/AuthenticationView/AuthenticationView';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import AuthenticationLevel from '../../../types/AuthenticationLevel';
import FetchStateBehavior from '../../../behaviors/FetchStateBehavior';

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

const mapStateToProps = (state: RootState, ownProps: OwnProps): StateProps => {
  const stage = (state.authentication.remoteState)
    ? authenticationLevelToStage(state.authentication.remoteState.authentication_level)
    : Stage.FIRST_FACTOR;

  let url: string | null = null;
  if (ownProps.location) {
    const params = QueryString.parse(ownProps.location.search);
    if ('rd' in params) {
      url = params['rd'] as string;
    } else if (state.authentication.remoteState && state.authentication.remoteState.default_redirection_url) {
      url = state.authentication.remoteState.default_redirection_url;
    }
  }

  return {
    redirectionUrl: url,
    remoteState: state.authentication.remoteState,
    stage: stage,
  };
}

const mapDispatchToProps = (dispatch: Dispatch) => {
  return {
    onInit: async () => await FetchStateBehavior(dispatch)
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(AuthenticationView);