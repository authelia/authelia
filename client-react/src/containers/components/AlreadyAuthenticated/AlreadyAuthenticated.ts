import { connect } from 'react-redux';
import { Dispatch } from 'redux';
import { RootState } from '../../../reducers';
import AlreadyAuthenticated, { DispatchProps } from '../../../components/AlreadyAuthenticated/AlreadyAuthenticated';
import LogoutBehavior from '../../../behaviors/LogoutBehavior';

const mapStateToProps = (state: RootState) => {
  return {};
}

const mapDispatchToProps = (dispatch: Dispatch): DispatchProps => {
  return {
    onLogoutClicked: () => LogoutBehavior(dispatch),
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(AlreadyAuthenticated);