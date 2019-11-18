import React, { ReactNode } from "react";
import { makeStyles } from "@material-ui/core";
import classnames from "classnames";

interface IconWithContextProps {
    icon: ReactNode;
    context: ReactNode;

    className?: string;
}

export default function (props: IconWithContextProps) {
    const iconSize = 64;
    const style = makeStyles(theme => ({
        root: {
            height: iconSize + theme.spacing(6),
        },
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
            height: theme.spacing(6),
        }
    }))();

    return (
        <div className={classnames(props.className, style.root)}>
            <div className={style.iconContainer}>
                <div className={style.icon}>
                    {props.icon}
                </div>
            </div>
            <div className={style.context}>
                {props.context}
            </div>
        </div>
    )
}