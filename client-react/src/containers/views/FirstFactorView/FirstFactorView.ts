import { connect } from 'react-redux';
import FirstFactorView, { Props } from '../../../views/FirstFactorView/FirstFactorView';
import { Dispatch } from 'redux';
import { authenticateFailure, authenticateSuccess, authenticate } from '../../../reducers/Portal/actions';
import { RootState } from '../../../reducers';


const mapStateToProps = (state: RootState) => ({});

function onAuthenticationRequested(dispatch: Dispatch, ownProps: Props) {
  return async (username: string, password: string) => {
    dispatch(authenticate());
    fetch('/api/firstfactor', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: username,
        password: password,
      })
    }).then(async (res) => {
      const json = await res.json();
      if ('error' in json) {
        dispatch(authenticateFailure(json['error']));
        return;
      }
      dispatch(authenticateSuccess());
      ownProps.history.push('/2fa');
    });
  }
}

const mapDispatchToProps = (dispatch: Dispatch, ownProps: Props) => {
  return {
    onAuthenticationRequested: onAuthenticationRequested(dispatch, ownProps),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(FirstFactorView);