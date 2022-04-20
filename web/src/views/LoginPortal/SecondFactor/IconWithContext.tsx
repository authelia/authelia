import React, { ReactNode } from "react";

import { Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import classnames from "classnames";

interface IconWithContextProps {
    icon: ReactNode;
    children: ReactNode;

    className?: string;
}

const IconWithContext = function (props: IconWithContextProps) {
    const iconSize = 64;
    const styles = makeStyles((theme: Theme) => ({
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
        <div className={classnames(props.className, styles.root)}>
            <div className={styles.iconContainer}>
                <div className={styles.icon}>{props.icon}</div>
            </div>
            <div className={styles.context}>{props.children}</div>
        </div>
    );
};

export default IconWithContext;
