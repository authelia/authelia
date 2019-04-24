import { connect } from 'react-redux';
import { RootState } from '../../../reducers';
import { Dispatch } from 'redux';
import SecondFactorDuoPush, { StateProps, OwnProps, DispatchProps } from '../../../components/SecondFactorDuoPush/SecondFactorDuoPush';
import FetchStateBehavior from '../../../behaviors/FetchStateBehavior';
import TriggerDuoPushAuth from '../../../behaviors/TriggerDuoPushAuth';


const mapStateToProps = (state: RootState): StateProps => ({
  duoPushVerified: state.secondFactor.duoPushVerificationSuccess,
  duoPushError: state.secondFactor.duoPushVerificationError,
});

async function redirectIfPossible(body: any) {
  if (body && 'redirect' in body) {
    window.location.href = body['redirect'];
    return true;
  }
  return false;
}

async function handleSuccess(dispatch: Dispatch, body: {redirect: string} | undefined, duration?: number) {
  async function handle() {
    const redirected = await redirectIfPossible(body);
    if (!redirected) {
      await FetchStateBehavior(dispatch);
    }
  }

  if (duration) {
    setTimeout(handle, duration);
  } else {
    await handle();
  }
}

async function triggerDuoPushAuth(dispatch: Dispatch, redirectionUrl: string | null) {
  const body = await TriggerDuoPushAuth(dispatch, redirectionUrl);
  await handleSuccess(dispatch, body, 1000);
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: OwnProps): DispatchProps => {
  return {
    onInit: async () => {
      await triggerDuoPushAuth(dispatch, ownProps.redirectionUrl);
    },
    onRetryClicked: async () => {
      await triggerDuoPushAuth(dispatch, ownProps.redirectionUrl);
    }
  }
}


export default connect(mapStateToProps, mapDispatchToProps)(SecondFactorDuoPush);