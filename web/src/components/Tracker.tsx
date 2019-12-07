import React, { useEffect, useCallback, Fragment, ReactNode } from "react";
import { useLocation } from "react-router";
import ReactGA, { Tracker } from "react-ga";

export interface Props {
    tracker: Tracker | undefined;
    children?: ReactNode;
}

export default function (props: Props) {
    const location = useLocation();

    const trackPage = useCallback((page: string) => {
        if (props.tracker) {
            ReactGA.set({ page });
            ReactGA.pageview(page);
        }
    }, [props.tracker]);

    useEffect(() => {
        if (props.tracker) {
            ReactGA.initialize([props.tracker]);
        }
    }, [props.tracker]);

    useEffect(() => {
        trackPage(location.pathname + location.search);
    }, [trackPage, location.pathname, location.search]);

    return (
        <Fragment>
            {props.children}
        </Fragment>
    )
}