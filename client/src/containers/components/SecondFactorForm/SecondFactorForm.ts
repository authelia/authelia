import { connect } from 'react-redux';
import { Dispatch } from 'redux';
import SecondFactorForm from '../../../components/SecondFactorForm/SecondFactorForm';
import LogoutBehavior from '../../../behaviors/LogoutBehavior';
import { RootState } from '../../../reducers';
import { StateProps, DispatchProps } from '../../../components/SecondFactorForm/SecondFactorForm';
import FetchPrefered2faMethod from '../../../behaviors/FetchPrefered2faMethod';
import { setUseAnotherMethod, setSecurityKeySupported } from '../../../reducers/Portal/SecondFactor/actions';
import GetAvailable2faMethods from '../../../behaviors/GetAvailable2faMethods';
import u2fApi from 'u2f-api';


const mapStateToProps = (state: RootState): StateProps => {
  return {
    method: state.secondFactor.preferedMethod,
    useAnotherMethod: state.secondFactor.userAnotherMethod,
  }
}

const mapDispatchToProps = (dispatch: Dispatch): DispatchProps => {
  return {
    onInit: async () => {
      dispatch(setSecurityKeySupported(await u2fApi.isSupported()));
      FetchPrefered2faMethod(dispatch);
      GetAvailable2faMethods(dispatch);
    },
    onLogoutClicked: () => LogoutBehavior(dispatch),
    onUseAnotherMethodClicked: () => dispatch(setUseAnotherMethod(true)),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(SecondFactorForm);