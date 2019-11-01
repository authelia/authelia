import React from "react"
import { Redirect } from "react-router";

export interface DispatchProps {
    onInit: () => void;
}

type Props = DispatchProps;

export default class LogoutView extends React.Component<Props> {
    componentWillMount() {
        this.props.onInit();
    }

    render() {
        return <Redirect to='/' />;
    }
}