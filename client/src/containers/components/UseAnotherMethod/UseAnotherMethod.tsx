import { connect } from 'react-redux';
import { Dispatch } from 'redux';
import { RootState } from '../../../reducers';
import SetPrefered2faMethod from '../../../behaviors/SetPrefered2faMethod';
import { getPreferedMethodSuccess, setUseAnotherMethod } from '../../../reducers/Portal/SecondFactor/actions';
import Method2FA from '../../../types/Method2FA';
import UseAnotherMethod, {StateProps, DispatchProps} from '../../../components/UseAnotherMethod/UseAnotherMethod';
import GetAvailable2faMethods from '../../../behaviors/GetAvailable2faMethods';

const mapStateToProps = (state: RootState): StateProps => ({
  availableMethods: state.secondFactor.getAvailableMethodResponse,
})

async function storeMethod(dispatch: Dispatch, method: Method2FA) {
  // display the new option
  dispatch(getPreferedMethodSuccess(method));
  dispatch(setUseAnotherMethod(false));

  // And save the method for next time.
  await SetPrefered2faMethod(dispatch, method);
}

const mapDispatchToProps = (dispatch: Dispatch): DispatchProps => {
  return {
    onInit: () => GetAvailable2faMethods(dispatch),
    onOneTimePasswordMethodClicked: () => storeMethod(dispatch, 'totp'),
    onSecurityKeyMethodClicked: () => storeMethod(dispatch, 'u2f'),
    onDuoPushMethodClicked: () => storeMethod(dispatch, "duo_push"),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(UseAnotherMethod);