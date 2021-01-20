import React, { ReactNode } from "react";

import { makeStyles } from "@material-ui/core";
import classnames from "classnames";

interface IconWithContextProps {
    icon: ReactNode;
    context: ReactNode;

    className?: string;
}

const IconWithContext = function (props: IconWithContextProps) {
    const iconSize = 64;
    const style = makeStyles((theme) => ({
        root: {},
        iconContainer: {
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
        },
        icon: {
            width: iconSize,
            height: iconSize,
        },
        context: {
            display: "block",
        },
    }))();

    return (
        <div className={classnames(props.className, style.root)}>
            <div className={style.iconContainer}>
                <div className={style.icon}>{props.icon}</div>
            </div>
            <div className={style.context}>{props.context}</div>
        </div>
    );
};

export default IconWithContext;
