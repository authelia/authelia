import { connect } from 'react-redux';
import { Dispatch } from 'redux';
import SecondFactorForm from '../../../components/SecondFactorForm/SecondFactorForm';
import LogoutBehavior from '../../../behaviors/LogoutBehavior';
import { RootState } from '../../../reducers';
import { StateProps, DispatchProps } from '../../../components/SecondFactorForm/SecondFactorForm';
import FetchPrefered2faMethod from '../../../behaviors/FetchPrefered2faMethod';
import SetPrefered2faMethod from '../../../behaviors/SetPrefered2faMethod';
import { getPreferedMethodSuccess, setUseAnotherMethod } from '../../../reducers/Portal/SecondFactor/actions';
import Method2FA from '../../../types/Method2FA';

const mapStateToProps = (state: RootState): StateProps => {
  return {
    method: state.secondFactor.preferedMethod,
    useAnotherMethod: state.secondFactor.userAnotherMethod,
  }
}

async function storeMethod(dispatch: Dispatch, method: Method2FA) {
  // display the new option
  dispatch(getPreferedMethodSuccess(method));
  dispatch(setUseAnotherMethod(false));

  // And save the method for next time.
  await SetPrefered2faMethod(dispatch, method);
}

const mapDispatchToProps = (dispatch: Dispatch): DispatchProps => {
  return {
    onInit: () => FetchPrefered2faMethod(dispatch),
    onLogoutClicked: () => LogoutBehavior(dispatch),
    onOneTimePasswordMethodClicked: () => storeMethod(dispatch, 'totp'),
    onSecurityKeyMethodClicked: () => storeMethod(dispatch, 'u2f'),
    onDuoPushMethodClicked: () => storeMethod(dispatch, "duo_push"),
    onUseAnotherMethodClicked: () => dispatch(setUseAnotherMethod(true)),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(SecondFactorForm);