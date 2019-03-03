import { connect } from 'react-redux';
import PortalLayout from '../../../layouts/PortalLayout/PortalLayout';
import { RootState } from '../../../reducers';

const mapStateToProps = (state: RootState) => ({});

export default connect(mapStateToProps)(PortalLayout);